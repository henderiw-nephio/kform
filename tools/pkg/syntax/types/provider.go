package types

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
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/sets"
	"github.com/henderiw/logger/log"
)

type Provider struct {
	NSN      cache.NSN
	ExecPath string
	Initializer
	Resources       sets.Set[string]
	ReadDataSources sets.Set[string]
	ListDataSources sets.Set[string]
}

type Initializer func() (kfplugin.Provider, error)

func (r *Provider) Init(ctx context.Context, rootPath string, nsn cache.NSN, reqs []v1alpha1.Provider) error {
	log := log.FromContext(ctx)
	providerPath := filepath.Join(rootPath, ".kform", "providers")
	p, err := address.GetPackage(nsn, reqs)
	if err != nil {
		return err
	}
	r.NSN = nsn
	r.ExecPath = filepath.Join(providerPath, p.Path())
	r.Initializer = ProviderInitializer(r.ExecPath)
	r.Resources = sets.New[string]()
	r.ReadDataSources = sets.New[string]()
	r.ListDataSources = sets.New[string]()
	// initialize the provider
	provider, err := r.Initializer()
	if err != nil {
		log.Error("failed starting provider", "nsn", r.NSN.Name)
		return fmt.Errorf("failed starting provider %s, err: %s", r.NSN.Name, err.Error())
	}
	defer provider.Close(ctx)
	capResp, err := provider.Capabilities(ctx, &kfplugin1.Capabilities_Request{})
	if err != nil {
		log.Error("cannot get provider capabilities", "nsn", r.NSN.Name)
		return fmt.Errorf("cannot get provider capabilities %s, err: %s", r.NSN.Name, err.Error())
	}

	if len(capResp.Resources) > 0 {
		log.Info("resources", "nsn", r.NSN.Name, "resources", capResp.Resources)
		r.Resources.Insert(capResp.Resources...)
	}
	if len(capResp.ReadDataSources) > 0 {
		log.Info("read data sources", "nsn", r.NSN.Name, "resources", capResp.ReadDataSources)
		r.ReadDataSources.Insert(capResp.ReadDataSources...)
	}
	if len(capResp.ListDataSources) > 0 {
		log.Info("list data sources", "nsn", r.NSN.Name, "resources", capResp.ListDataSources)
		r.ListDataSources.Insert(capResp.ListDataSources...)
	}
	return nil
}

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
