package main

import (
	"log/slog"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1"
	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfserver1"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/schema"
	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes"
	"github.com/henderiw/logger/log"
)

const providerName = "registry.fkorm.io/kform/kubernetes"

func main() {

	l := log.NewLogger(&log.HandlerOptions{Name: "kubernetes-provider-logger", AddSource: false})
	slog.SetDefault(l)

	grpcProviderFunc := func() kfprotov1.ProviderServer {
		return schema.NewGRPCProviderServer(kubernetes.Provider())
	}

	opts := []kfserver1.ServeOpt{
		kfserver1.WithGoPluginLogger(l),
	}
	if err := kfserver1.Serve(
		providerName,
		//func() kfprotov1.ProviderServer {
		//	return schema.NewGRPCProviderServer(kubernetes.Provider())
		//},
		grpcProviderFunc,
		opts...); err != nil {
		slog.Error("kform serve failed", "err", err)
	}
	slog.Info("done serving kform")

	/*
		plugin.Serve(&plugin.ServeConfig{
			HandshakeConfig: kfplugin.Handshake,
			VersionedPlugins: map[int]plugin.PluginSet{
				kfplugin.DefaultProviderPluginVersion: {
					kfplugin.ProviderPluginName: &kfserver1.GRPCProviderPlugin{
						GRPCProvider: grpcProviderFunc,
						Name:         providerName,
					}},
			},
			// A non-nil value here enables gRPC serving for this plugin...
			GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
				return grpc.NewServer(append(opts,
					grpc.MaxSendMsgSize(64<<20 ),
					grpc.MaxRecvMsgSize(64<<20 ))...)
			},
		})
	*/
}
