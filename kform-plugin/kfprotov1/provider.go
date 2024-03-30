package kfprotov1

import (
	"context"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
)

type ProviderServer interface {
	Capabilities(ctx context.Context, in *kfplugin1.Capabilities_Request) (*kfplugin1.Capabilities_Response, error)
	Configure(ctx context.Context, in *kfplugin1.Configure_Request) (*kfplugin1.Configure_Response, error)
	StopProvider(ctx context.Context, in *kfplugin1.StopProvider_Request) (*kfplugin1.StopProvider_Response, error)

	DataSourceServer
	ResourceServer
}

type DataSourceServer interface {
	ReadDataSource(ctx context.Context, in *kfplugin1.ReadDataSource_Request) (*kfplugin1.ReadDataSource_Response, error)
	ListDataSource(ctx context.Context, in *kfplugin1.ListDataSource_Request) (*kfplugin1.ListDataSource_Response, error)
}

type ResourceServer interface {
	ReadResource(ctx context.Context, in *kfplugin1.ReadResource_Request) (*kfplugin1.ReadResource_Response, error)
	CreateResource(ctx context.Context, in *kfplugin1.CreateResource_Request) (*kfplugin1.CreateResource_Response, error)
	UpdateResource(ctx context.Context, in *kfplugin1.UpdateResource_Request) (*kfplugin1.UpdateResource_Response, error)
	DeleteResource(ctx context.Context, in *kfplugin1.DeleteResource_Request) (*kfplugin1.DeleteResource_Response, error)
}
