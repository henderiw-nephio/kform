package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/henderiw-nephio/kform/plugin"
	"github.com/henderiw-nephio/kform/plugin/examples/grpc/shared"
	"github.com/henderiw/logger/log"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	l := log.NewLogger(&log.HandlerOptions{Name: "example-kv-plugin", AddSource: false})
	slog.SetDefault(l)

	ctx := ctrl.SetupSignalHandler()
	ctx = log.IntoContext(ctx, l)

	log := l

	// We're a host. Start by launching the plugin process.
	cmdStr := "docker run --network host europe-docker.pkg.dev/srlinux/eu.gcr.io/provider-grpc-example:latest"
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		VersionedPlugins: map[int]plugin.PluginSet{
			1: shared.PluginMap,
		},
		//Cmd: exec.Command("sh", "-c", os.Getenv("KV_PLUGIN")),
		Cmd: exec.Command("/bin/sh", "-c", cmdStr),
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Error("cannot create rpc client", "error", err.Error())
		os.Exit(1)
	}
	// Request the plugin
	raw, err := rpcClient.Dispense("kv_grpc")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	log.Info("client dispense", "kv", raw)

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	kv := raw.(shared.KV)
	os.Args = os.Args[1:]
	switch os.Args[0] {
	case "get":
		result, err := kv.Get(os.Args[1])
		if err != nil {
			fmt.Println("Error:", err.Error())
			l.Error("cannot get value", "args", os.Args, "error", "")
			os.Exit(1)
		}

		fmt.Println(string(result))

	case "put":
		err := kv.Put(os.Args[1], []byte(os.Args[2]))
		if err != nil {
			fmt.Println("Error:", err.Error())
			os.Exit(1)
		}

	default:
		fmt.Printf("Please only use 'get' or 'put', given: %q", os.Args[0])
		os.Exit(1)
	}
	os.Exit(0)
}
