// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/henderiw-nephio/kform/plugin"
	"github.com/henderiw-nephio/kform/plugin/examples/grpc/shared"
	"google.golang.org/grpc"
)

// Here is a real implementation of KV that writes to a local file with
// the key name and the contents are the value of the key.
type KV struct{}

func (KV) Put(key string, value []byte) error {
	value = []byte(fmt.Sprintf("%s\n\nWritten from plugin-go-grpc", string(value)))
	return os.WriteFile("kv_"+key, value, 0644)
}

func (KV) Get(key string) ([]byte, error) {
	return os.ReadFile("kv_" + key)
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		VersionedPlugins: map[int]plugin.PluginSet{
			1: {"kv": &shared.KVGRPCPlugin{Impl: &KV{}}},
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			return grpc.NewServer(append(opts,
				grpc.MaxSendMsgSize(64<<20 /* 64MB */),
				grpc.MaxRecvMsgSize(64<<20 /* 64MB */))...)
		},
	})
}
