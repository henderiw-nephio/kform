package plugin

import (
	"context"

	"google.golang.org/grpc"
)

// Plugin is the interface that is implemented to serve/connect to
// a plugin over gRPC.
type Plugin interface {
	// GRPCServer should register this plugin for serving with the
	// given GRPCServer. Unlike Plugin.Server, this is only called once
	// since gRPC plugins serve singletons.
	GRPCServer(*GRPCBroker, *grpc.Server) error

	// GRPCClient should return the interface implementation for the plugin
	// you're serving via gRPC. The provided context will be canceled by
	// go-plugin in the event of the plugin process exiting.
	GRPCClient(context.Context, *GRPCBroker, *grpc.ClientConn) (interface{}, error)
}
