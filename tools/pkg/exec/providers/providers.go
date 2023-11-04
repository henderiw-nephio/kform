package providers

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
	kfplugin "github.com/henderiw-nephio/kform/kform-plugin/plugin"
	"github.com/henderiw-nephio/kform/plugin"
	"github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/providers/logging"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/sets"
	"github.com/henderiw/logger/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Initialize(ctx context.Context, rootPath string, provReqs map[cache.NSN][]v1alpha1.Provider) (Providers, error) {
	log := log.FromContext(ctx)
	providers := cache.New[Provider]()

	for nsn := range provReqs {
		execPath := filepath.Join(rootPath, ".kform/providers", fmt.Sprintf("provider-%s", nsn.Name))

		p := Provider{
			Initializer:     ProviderInitializer(execPath),
			GVK:             schema.GroupVersionKind{},
			Resources:       sets.New[string](),
			ReadDataSources: sets.New[string](),
			ListDataSources: sets.New[string](),
		}

		// initialize the provider
		provider, err := p.Initializer()
		if err != nil {
			log.Error("failed starting provider", "nsn", nsn)
			return nil, fmt.Errorf("failed starting provider %s, err: %s", nsn, err.Error())
		}
		defer provider.Close(ctx)
		capResp, err := provider.Capabilities(ctx, &kfplugin1.Capabilities_Request{})
		if err != nil {
			log.Error("cannot get provider capabilities", "nsn", nsn)
			return nil, fmt.Errorf("cannot get provider capabilities %s, err: %s", nsn, err.Error())
		}

		if len(capResp.Resources) > 0 {
			log.Info("resources", "nsn", nsn, "resources", capResp.Resources)
			p.Resources.Insert(capResp.Resources...)
		}
		if len(capResp.ReadDataSources) > 0 {
			log.Info("read data sources", "nsn", nsn, "resources", capResp.ReadDataSources)
			p.Resources.Insert(capResp.ReadDataSources...)
		}
		if len(capResp.ListDataSources) > 0 {
			log.Info("list data sources", "nsn", nsn, "resources", capResp.ListDataSources)
			p.Resources.Insert(capResp.ListDataSources...)
		}
		providers.Add(ctx, nsn, p)
	}
	return providers, nil
}

type Provider struct {
	Initializer
	GVK             schema.GroupVersionKind
	Resources       sets.Set[string]
	ReadDataSources sets.Set[string]
	ListDataSources sets.Set[string]
}

type Providers cache.Cache[Provider]

type Initializer func() (kfplugin.Provider, error)

// ProviderInitializer produces a provider factory that runs up the executable
// file in the given path and uses go-plugin to implement
// Provider Interface against it.
func ProviderInitializer(execPath string) Initializer {
	return func() (kfplugin.Provider, error) {

		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig:  kfplugin.Handshake,
			VersionedPlugins: kfplugin.VersionedPlugins,
			//AutoMTLS:         enableProviderAutoMTLS,
			Cmd: exec.Command(execPath),
			//Cmd:        exec.Command("./bin/provider-kubernetes"),
			SyncStdout: logging.PluginOutputMonitor(fmt.Sprintf("%s:stdout", "test")),
			SyncStderr: logging.PluginOutputMonitor(fmt.Sprintf("%s:stderr", "test")),
		})

		// Connect via RPC
		rpcClient, err := client.Client()
		if err != nil {
			return nil, err
		}

		// Request the plugin
		raw, err := rpcClient.Dispense(kfplugin.ProviderPluginName)
		if err != nil {
			return nil, err
		}

		p := raw.(*kfplugin.GRPCProvider)
		p.PluginClient = client
		return p, nil
	}
}
