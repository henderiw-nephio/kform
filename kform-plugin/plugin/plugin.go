package plugin

import (
	"github.com/henderiw-nephio/kform/plugin"
)

// Handshake is the HandshakeConfig used to configure clients and servers.
var Handshake = plugin.HandshakeConfig{
	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "KF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

const ProviderPluginName = "provider"
const DefaultProviderPluginVersion = 1

// VersionedPlugins includes both protocol 5 and 6 because this is the function
// called in providerFactory (command/meta_providers.go) to set up the initial
// plugin client config.
var VersionedPlugins = map[int]plugin.PluginSet{
	DefaultProviderPluginVersion: {
		ProviderPluginName: &GRPCProviderPlugin{},
	},
}
