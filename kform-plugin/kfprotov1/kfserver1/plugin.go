package kfserver1

import (
	"context"
	"errors"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1"
	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
	"github.com/henderiw-nephio/kform/plugin"
	"google.golang.org/grpc"
)

type GRPCProviderPlugin struct {
	GRPCProvider func() kfprotov1.ProviderServer
	Opts         []ServeOpt
	Name         string
}

func (p *GRPCProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	kfplugin1.RegisterProviderServer(s, New(p.Name, p.GRPCProvider(), p.Opts...))
	return nil
}

func (p *GRPCProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return nil, errors.New("kform-plugin-go only implements gRPC servers")
}
