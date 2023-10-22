package plugin

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"os/user"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/henderiw/logger/log"
	"google.golang.org/grpc"
)

// CoreProtocolVersion is the ProtocolVersion of the plugin system itself.
// We will increment this whenever we change any protocol behavior. This
// will invalidate any prior plugins but will at least allow us to iterate
// on the core in a safe way. We will do our best to do this very
// infrequently.
const CoreProtocolVersion = 1

// HandshakeConfig is the configuration used by client and servers to
// handshake before starting a plugin connection. This is embedded by
// both ServeConfig and ClientConfig.
//
// In practice, the plugin host creates a HandshakeConfig that is exported
// and plugins then can easily consume it.
type HandshakeConfig struct {
	// MagicCookieKey and value are used as a very basic verification
	// that a plugin is intended to be launched. This is not a security
	// measure, just a UX feature. If the magic cookie doesn't match,
	// we show human-friendly output.
	MagicCookieKey   string
	MagicCookieValue string
}

// PluginSet is a set of plugins provided to be registered in the plugin
// server.
type PluginSet map[string]Plugin

// ServeConfig configures what sorts of plugins are served.
type ServeConfig struct {
	// HandshakeConfig is the configuration that must match clients.
	HandshakeConfig

	// TLSProvider is a function that returns a configured tls.Config.
	TLSProvider func() (*tls.Config, error)

	// VersionedPlugins is a map of PluginSets for specific protocol versions.
	// These can be used to negotiate a compatible version between client and
	// server. If this is set, Handshake.ProtocolVersion is not required.
	VersionedPlugins map[int]PluginSet

	// GRPCServer should be non-nil to enable serving the plugins over
	// gRPC. This is a function to create the server when needed with the
	// given server options. The server options populated by go-plugin will
	// be for TLS if set. You may modify the input slice.
	//
	// Note that the grpc.Server will automatically be registered with
	// the gRPC health checking service. This is not optional since go-plugin
	// relies on this to implement Ping().
	GRPCServer func([]grpc.ServerOption) *grpc.Server

	// Logger is used to pass a logger into the server. If none is provided the
	// server will create a default logger.
	Logger *slog.Logger
}

// Serve serves the plugins given by ServeConfig.
//
// Serve doesn't return until the plugin is done being executed. Any
// fixable errors will be output to os.Stderr and the process will
// exit with a status code of 1. Serve will panic for unexpected
// conditions where a user's fix is unknown.
//
// This is the method that plugins should call in their main() functions.
func Serve(opts *ServeConfig) {
	exitCode := -1
	// We use this to trigger an `os.Exit` so that we can execute our other
	// deferred functions. In test mode, we just output the err to stderr
	// and return.
	defer func() {
		if exitCode >= 0 {
			os.Exit(exitCode)
		}
	}()

	// Validate the handshake config
	if opts.MagicCookieKey == "" || opts.MagicCookieValue == "" {
		fmt.Fprintf(os.Stderr,
			`cannot serve this plugin: no magic cookie key or value was set`)
		exitCode = 1
		return
	}
	if os.Getenv(opts.MagicCookieKey) != opts.MagicCookieValue {
		fmt.Fprintf(os.Stderr,
			`cannot execute this plugin direct, execute the plugin via the plugin loader`)
		exitCode = 1
		return
	}

	l := opts.Logger
	if l == nil {
		// internal logger to os.Stderr
		l = log.NewLogger(&log.HandlerOptions{Name: "plugin", AddSource: false})
	}

	// negotiate the version and plugins
	// start with default version in the handshake config
	protoVersion, pluginSet, err := protocolVersion(opts)
	if err != nil {
		l.Error("cannot initialize plugin", "error", err)
		return
	}

	// Register a listener so we can accept a connection
	listener, err := serverListener(unixSocketConfigFromEnv())
	if err != nil {
		l.Error("cannot initialize plugin", "error", err)
		return
	}

	// Close the listener on return. We wrap this in a func() on purpose
	// because the "listener" reference may change to TLS.
	defer func() {
		listener.Close()
	}()

	var tlsConfig *tls.Config
	if opts.TLSProvider != nil {
		tlsConfig, err = opts.TLSProvider()
		if err != nil {
			l.Error("cannot initialize plugin tls", "error", err)
			return
		}
	}

	var serverCert string
	clientCert := os.Getenv("PLUGIN_CLIENT_CERT")
	// If the client is configured using AutoMTLS, the certificate will be here,
	// and we need to generate our own in response.
	if tlsConfig == nil && clientCert != "" {
		l.Info("configuring server automatic mTLS")
		clientCertPool := x509.NewCertPool()
		if !clientCertPool.AppendCertsFromPEM([]byte(clientCert)) {
			l.Error("client cert provided but failed to parse", "cert", clientCert)
		}

		certPEM, keyPEM, err := generateCert()
		if err != nil {
			l.Error("failed to generate server certificate", "error", err)
			panic(err)
		}

		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			l.Error("failed to parse server certificate", "error", err)
			panic(err)
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clientCertPool,
			MinVersion:   tls.VersionTLS12,
			RootCAs:      clientCertPool,
			ServerName:   "localhost",
		}

		// We send back the raw leaf cert data for the client rather than the
		// PEM, since the protocol can't handle newlines.
		serverCert = base64.RawStdEncoding.EncodeToString(cert.Certificate[0])
	}

	// Create the channel to tell us when we're done
	doneCh := make(chan struct{})

	// Create our new stdout, stderr files. These will override our built-in
	// stdout/stderr so that it works across the stream boundary.
	var stdout_r, stderr_r io.Reader
	stdout_r, stdout_w, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing plugin: %s\n", err)
		os.Exit(1)
	}
	stderr_r, stderr_w, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing plugin: %s\n", err)
		os.Exit(1)
	}

	server := &GRPCServer{
		Plugins: pluginSet,
		Server:  opts.GRPCServer,
		TLS:     tlsConfig,
		Stdout:  stdout_r,
		Stderr:  stderr_r,
		DoneCh:  doneCh,
		logger:  l,
	}

	// Initialize the servers
	if err := server.Init(); err != nil {
		l.Error("cannot initialize protocol", "error", err)
		return
	}

	l.Debug("plugin address",
		"network", listener.Addr().Network(),
		"address", listener.Addr().String(),
	)

	fmt.Printf("%d|%d|%s|%s|%s\n",
		CoreProtocolVersion,
		protoVersion,
		listener.Addr().Network(),
		listener.Addr().String(),
		serverCert)
	os.Stdout.Sync()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		count := 0
		for {
			<-ch
			count++
			l.Info("plugin received interrupt signal, ignoring", "count", count)
		}
	}()

	os.Stdout = stdout_w
	os.Stderr = stderr_w

	// Accept connections and wait for completion
	go server.Serve(listener)

	select {
	case <-doneCh:
		// Note that given the documentation of Serve we should probably be
		// setting exitCode = 0 and using os.Exit here. That's how it used to
		// work before extracting this library. However, for years we've done
		// this so we'll keep this functionality.
	}
}

func unixSocketConfigFromEnv() UnixSocketConfig {
	return UnixSocketConfig{
		Group:     os.Getenv(EnvUnixSocketGroup),
		socketDir: os.Getenv(EnvUnixSocketDir),
	}
}

// protocolVersion determines the protocol version and plugin set to be used by
// the server. In the event that there is no suitable version, the last version
// in the config is returned leaving the client to report the incompatibility.
func protocolVersion(opts *ServeConfig) (int, PluginSet, error) {
	var clientVersions []int
	if vs := os.Getenv("PLUGIN_PROTOCOL_VERSIONS"); vs != "" {
		for _, s := range strings.Split(vs, ",") {
			v, err := strconv.Atoi(s)
			if err != nil {
				fmt.Fprintf(os.Stderr, "server sent invalid plugin version %q", s)
				continue
			}
			clientVersions = append(clientVersions, v)
		}
	}

	// Sort the version to make sure we match the latest first
	var versions []int
	for v := range opts.VersionedPlugins {
		versions = append(versions, v)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(versions)))

	// See if we have multiple versions of Plugins to choose from
	for _, version := range versions {
		version := version

		for _, clientVersion := range clientVersions {
			if clientVersion == version {
				return version, opts.VersionedPlugins[version], nil
			}
		}
	}

	return 0, nil, fmt.Errorf("no matching protocol version found, clientVersions %v, version: %v", clientVersions, versions)
}

func serverListener(unixSocketCfg UnixSocketConfig) (net.Listener, error) {
	if runtime.GOOS == "windows" {
		return serverListener_tcp()
	}

	return serverListener_unix(unixSocketCfg)
}

func serverListener_tcp() (net.Listener, error) {
	envMinPort := os.Getenv("PLUGIN_MIN_PORT")
	envMaxPort := os.Getenv("PLUGIN_MAX_PORT")

	var minPort, maxPort int64
	var err error

	switch {
	case len(envMinPort) == 0:
		minPort = 0
	default:
		minPort, err = strconv.ParseInt(envMinPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("cannot get value from PLUGIN_MIN_PORT: %v", err)
		}
	}

	switch {
	case len(envMaxPort) == 0:
		maxPort = 0
	default:
		maxPort, err = strconv.ParseInt(envMaxPort, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("cannot get value from PLUGIN_MAX_PORT: %v", err)
		}
	}

	if minPort > maxPort {
		return nil, fmt.Errorf("PLUGIN_MIN_PORT value of %d is greater than PLUGIN_MAX_PORT value of %d", minPort, maxPort)
	}

	for port := minPort; port <= maxPort; port++ {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			return listener, nil
		}
	}

	return nil, errors.New("cannot bind plugin TCP listener")
}

func serverListener_unix(unixSocketCfg UnixSocketConfig) (net.Listener, error) {
	tf, err := os.CreateTemp(unixSocketCfg.socketDir, "plugin")
	if err != nil {
		return nil, err
	}
	path := tf.Name()

	// Close the file and remove it because it has to not exist for
	// the domain socket.
	if err := tf.Close(); err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}

	l, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	// By default, unix sockets are only writable by the owner. Set up a custom
	// group owner and group write permissions if configured.
	if unixSocketCfg.Group != "" {
		err = setGroupWritable(path, unixSocketCfg.Group, 0o660)
		if err != nil {
			return nil, err
		}
	}

	// Wrap the listener in rmListener so that the Unix domain socket file
	// is removed on close.
	return &rmListener{
		Listener: l,
		Path:     path,
	}, nil
}

func setGroupWritable(path, groupString string, mode os.FileMode) error {
	groupID, err := strconv.Atoi(groupString)
	if err != nil {
		group, err := user.LookupGroup(groupString)
		if err != nil {
			return fmt.Errorf("failed to find gid from %q: %w", groupString, err)
		}
		groupID, err = strconv.Atoi(group.Gid)
		if err != nil {
			return fmt.Errorf("failed to parse %q group's gid as an integer: %w", groupString, err)
		}
	}

	err = os.Chown(path, -1, groupID)
	if err != nil {
		return err
	}

	err = os.Chmod(path, mode)
	if err != nil {
		return err
	}

	return nil
}

// rmListener is an implementation of net.Listener that forwards most
// calls to the listener but also removes a file as part of the close. We
// use this to cleanup the unix domain socket on close.
type rmListener struct {
	net.Listener
	Path string
}

func (l *rmListener) Close() error {
	// Close the listener itself
	if err := l.Listener.Close(); err != nil {
		return err
	}

	// Remove the file
	return os.Remove(l.Path)
}
