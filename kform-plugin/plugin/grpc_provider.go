package plugin

import (
	"context"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
	"github.com/henderiw-nephio/kform/plugin"
	"google.golang.org/grpc"
)

// GRPCProviderPlugin implements plugin.GRPCPlugin for the go-plugin package.
type GRPCProviderPlugin struct {
	plugin.Plugin
	GRPCProvider func() kfplugin1.ProviderServer
}

func (p *GRPCProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCProvider{
		client: kfplugin1.NewProviderClient(c),
		ctx:    ctx,
	}, nil
}

func (p *GRPCProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	kfplugin1.RegisterProviderServer(s, p.GRPCProvider())
	return nil
}

// GRPCProvider handles the client, or core side of the plugin rpc connection.
// The GRPCProvider methods are mostly a translation layer between the
// terraform providers types and the grpc proto types, directly converting
// between the two.
type GRPCProvider struct {
	// PluginClient provides a reference to the plugin.Client which controls the plugin process.
	// This allows the GRPCProvider a way to shutdown the plugin process.
	PluginClient *plugin.Client

	// TestServer contains a grpc.Server to close when the GRPCProvider is being
	// used in an end to end test of a provider.
	//TestServer *grpc.Server

	// Addr uniquely identifies the type of provider.
	// Normally executed providers will have this set during initialization,
	// but it may not always be available for alternative execute modes.
	//Addr addrs.Provider

	// Proto client use to make the grpc service calls.
	client kfplugin1.ProviderClient

	// this context is created by the plugin package, and is canceled when the
	// plugin process ends.
	ctx context.Context

	// schema stores the schema for this provider. This is used to properly
	// serialize the requests for schemas.
	//mu     sync.Mutex
	//schema providers.GetProviderSchemaResponse
}

func (r *GRPCProvider) Capabilities(ctx context.Context, req *kfplugin1.Capabilities_Request) (*kfplugin1.Capabilities_Response, error) {
	return r.client.Capabilities(ctx, req)
}

func (r *GRPCProvider) Configure(ctx context.Context, req *kfplugin1.Configure_Request) (*kfplugin1.Configure_Response, error) {
	return r.client.Configure(ctx, req)
}
func (r *GRPCProvider) StopProvider(ctx context.Context, req *kfplugin1.StopProvider_Request) (*kfplugin1.StopProvider_Response, error) {
	return r.client.StopProvider(ctx, req)
}

func (r *GRPCProvider) ReadDataSource(ctx context.Context, req *kfplugin1.ReadDataSource_Request) (*kfplugin1.ReadDataSource_Response, error) {
	return r.client.ReadDataSource(ctx, req)
}
func (r *GRPCProvider) ListDataSource(ctx context.Context, req *kfplugin1.ListDataSource_Request) (*kfplugin1.ListDataSource_Response, error) {
	return r.client.ListDataSource(ctx, req)
}

func (r *GRPCProvider) ReadResource(ctx context.Context, req *kfplugin1.ReadResource_Request) (*kfplugin1.ReadResource_Response, error) {
	return r.client.ReadResource(ctx, req)
}
func (r *GRPCProvider) CreateResource(ctx context.Context, req *kfplugin1.CreateResource_Request) (*kfplugin1.CreateResource_Response, error) {
	return r.client.CreateResource(ctx, req)
}
