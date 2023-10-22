package shared

import (
	"context"

	"github.com/henderiw-nephio/kform/plugin/examples/grpc/proto"
)

// Here is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	proto.UnimplementedKVServer
	// This is the real implementation
	Impl KV
}

func (m *GRPCServer) Put(
	ctx context.Context,
	req *proto.PutRequest) (*proto.Empty, error) {
	return &proto.Empty{}, m.Impl.Put(req.Key, req.Value)
}

func (m *GRPCServer) Get(
	ctx context.Context,
	req *proto.GetRequest) (*proto.GetResponse, error) {
	v, err := m.Impl.Get(req.Key)
	return &proto.GetResponse{Value: v}, err
}
