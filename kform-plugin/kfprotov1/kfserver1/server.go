package kfserver1

import (
	"log/slog"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1"
	kfplugin "github.com/henderiw-nephio/kform/kform-plugin/plugin"
	"github.com/henderiw-nephio/kform/plugin"
	"google.golang.org/grpc"
)

// Kerraform

const (
	grpcMaxMessageSize = 256 << 20
)

type ServeOpt interface {
	ApplyServeOpt(*ServeConfig) error
}

type serveConfigFunc func(*ServeConfig) error

func (s serveConfigFunc) ApplyServeOpt(in *ServeConfig) error {
	return s(in)
}

type ServeConfig struct {
	logger *slog.Logger
}

// WithGoPluginLogger returns a ServeOpt that will set the logger that
// go-plugin should use to log messages.
func WithGoPluginLogger(logger *slog.Logger) ServeOpt {
	return serveConfigFunc(func(in *ServeConfig) error {
		in.logger = logger
		return nil
	})
}

func Serve(name string, serverFactory func() kfprotov1.ProviderServer, opts ...ServeOpt) error {
	conf := ServeConfig{}
	for _, opt := range opts {
		err := opt.ApplyServeOpt(&conf)
		if err != nil {
			return err
		}
	}

	serveConfig := &plugin.ServeConfig{
		HandshakeConfig: kfplugin.Handshake,
		VersionedPlugins: map[int]plugin.PluginSet{
			kfplugin.DefaultProviderPluginVersion: {
				kfplugin.ProviderPluginName: &GRPCProviderPlugin{
					Name:         name,
					Opts:         opts,
					GRPCProvider: serverFactory,
				}},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			opts = append(opts, grpc.MaxRecvMsgSize(grpcMaxMessageSize))
			opts = append(opts, grpc.MaxSendMsgSize(grpcMaxMessageSize))

			return grpc.NewServer(opts...)
		},
		Logger: conf.logger,
	}

	// in case of debug this becomes non blocking
	plugin.Serve(serveConfig)

	// TODO debug

	return nil
}
