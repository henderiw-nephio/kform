package shared

import (
	"context"

	"github.com/henderiw-nephio/kform/plugin/examples/grpc/proto"
)

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCClient struct{ client proto.KVClient }

func (m *GRPCClient) Put(key string, value []byte) error {
	_, err := m.client.Put(context.Background(), &proto.PutRequest{
		Key:   key,
		Value: value,
	})
	return err
}

func (m *GRPCClient) Get(key string) ([]byte, error) {
	resp, err := m.client.Get(context.Background(), &proto.GetRequest{
		Key: key,
	})
	if err != nil {
		return nil, err
	}

	return resp.Value, nil
}
