package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
	kfplugin "github.com/henderiw-nephio/kform/kform-plugin/plugin"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/plugin"
	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/api"
	"github.com/henderiw/logger/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	ctx := ctrl.SetupSignalHandler()

	l := log.NewLogger(&log.HandlerOptions{Name: "kform-logger", AddSource: false})
	slog.SetDefault(l)
	ctx = log.IntoContext(ctx, l)
	log := l

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  kfplugin.Handshake,
		VersionedPlugins: kfplugin.VersionedPlugins,
		//AutoMTLS:         enableProviderAutoMTLS,
		//SyncStdout:       logging.PluginOutputMonitor(fmt.Sprintf("%s:stdout", meta.Name)),
		//SyncStderr:       logging.PluginOutputMonitor(fmt.Sprintf("%s:stderr", meta.Name)),
		Cmd:        exec.Command("./bin/provider-kubernetes"),
		SyncStdout: PluginOutputMonitor(fmt.Sprintf("%s:stdout", "test")),
		SyncStderr: PluginOutputMonitor(fmt.Sprintf("%s:stderr", "test")),
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Error("cannot create rpc client", "error", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(kfplugin.ProviderPluginName)
	if err != nil {
		log.Error("cannot dispense rpc client", "error", err.Error())
		os.Exit(1)
	}

	log.Info("client dispense", "provider", raw)

	// store the client so that the plugin can kill the child process
	p := raw.(*kfplugin.GRPCProvider)
	p.PluginClient = client

	capResp, err := p.Capabilities(ctx, &kfplugin1.Capabilities_Request{})
	if err != nil {
		log.Error("cannot get capabilities", "error", err.Error())
		panic(err)
	}
	log.Info("capabilities response",
		"resources", capResp.Resources,
		"readDataSources", capResp.ReadDataSources,
		"listDataSources", capResp.ListDataSources,
		"diag", capResp.Diagnostics)

	conf := &api.ProviderAPI{
		Kind:      api.ProviderKindPackage,
		Directory: pointer.String("./examples/crd"),
	}
	confByte, err := json.Marshal(conf)
	if err != nil {
		log.Error("cannot json marshal config", "error", err.Error())
		panic(err)
	}

	confResp, err := p.Configure(ctx, &kfplugin1.Configure_Request{
		Config: confByte,
	})
	if err != nil {
		log.Error("cannot get capabilities", "error", err.Error())
		panic(err)
	}
	log.Info("configure response", "diag", confResp.Diagnostics)

	u := unstructured.Unstructured{}
	u.SetAPIVersion("apiextensions.k8s.io/v1")
	u.SetKind("CustomResourceDefinition")
	u.SetName("nodepools.inv.nephio.org")
	readByte, err := json.Marshal(&u)
	if err != nil {
		log.Error("cannot json marshal list", "error", err.Error())
		panic(err)
	}
	slog.Info("data", "req", string(readByte))

	readResp, err := p.ReadDataSource(ctx, &kfplugin1.ReadDataSource_Request{
		Name: "kubernetes_manifest",
		Data: readByte,
	})
	if err != nil {
		log.Error("cannot read resource", "error", err.Error())
		panic(err)
	}

	if err := json.Unmarshal(readResp.Data, &u); err != nil {
		log.Error("cannot unmarshal read resp", "error", err.Error())
		panic(err)
	}
	log.Info("read response",
		"apiVersion", u.GetAPIVersion(),
		"kind", u.GetKind(),
		"name", u.GetName(),
	)

	ul := unstructured.UnstructuredList{}
	ul.SetAPIVersion("apiextensions.k8s.io/v1")
	ul.SetKind("CustomResourceDefinition")
	listByte, err := json.Marshal(&ul)
	if err != nil {
		log.Error("cannot json marshal list", "error", err.Error())
		panic(err)
	}

	listResp, err := p.ListDataSource(ctx, &kfplugin1.ListDataSource_Request{
		Name: "kubernetes_manifest",
		Data: listByte,
	})
	if err != nil {
		log.Error("cannot get capabilities", "error", err.Error())
		panic(err)
	}

	if listResp.Diagnostics != nil && diag.Diagnostics(listResp.GetDiagnostics()).HasError() {
		log.Error("list failed", "error", diag.Diagnostics(listResp.GetDiagnostics()).Error())
		panic(diag.Diagnostics(listResp.GetDiagnostics()).HasError())
	}

	//log.Info("list response", "response", listResp)
	if listResp.GetData() != nil {
		if err := json.Unmarshal(listResp.GetData(), &ul); err != nil {
			log.Error("list failed", "error", diag.Diagnostics(listResp.GetDiagnostics()).Error())
			panic(err)
		}

		for _, u := range ul.Items {
			log.Info("list response",
				"apiVersion", u.GetAPIVersion(),
				"kind", u.GetKind(),
				"name", u.GetName(),
			)
		}
	}

	os.Exit(0)
}

// PluginOutputMonitor creates an io.Writer that will warn about any writes in
// the default logger. This is used to catch unexpected output from plugins,
// notifying them about the problem as well as surfacing the lost data for
// context.
func PluginOutputMonitor(source string) io.Writer {
	return pluginOutputMonitor{
		source: source,
		log:    log.NewLogger(&log.HandlerOptions{Name: "kform-plugin-logger", AddSource: false}),
	}
}

// pluginOutputMonitor is an io.Writer that logs all writes as
// "unexpected data" with the source name.
type pluginOutputMonitor struct {
	source string
	log    *slog.Logger
}

func (w pluginOutputMonitor) Write(d []byte) (int, error) {
	// Limit the write size to 1024 bytes We're not expecting any data to come
	// through this channel, so accidental writes will usually be stray fmt
	// debugging statements and the like, but we want to provide context to the
	// provider to indicate what the unexpected data might be.
	n := len(d)
	if n > 1024 {
		d = append(d[:1024], '.', '.', '.')
	}

	w.log.Warn("unexpected data", w.source, strings.TrimSpace(string(d)))
	return n, nil
}
