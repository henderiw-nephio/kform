package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

func main() {
	l := log.NewLogger(&log.HandlerOptions{Name: "test-reklease", AddSource: false})
	slog.SetDefault(l)

	// init context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx = log.IntoContext(ctx, l)
	log := log.FromContext(ctx)

	args := os.Args[1:]
	if len(args) < 2 {
		log.Error("cannt run command with less than 2 arguments", "got", args)
		os.Exit(1)
	}
	source := args[0]
	//version := args[1]

	split := strings.Split(source, "/")
	if len(split) != 3 {
		log.Error("arg[0] needs <hostname>/<org>/<namespace>", "got", source)
		os.Exit(1)
	}

	pkg, err := address.GetPackage(cache.NSN{Name: split[2]}, source)
	if err != nil {
		log.Error("cannot create new pkg", "err", err.Error())
		os.Exit(1)
	}
	if err := pkg.GetReleases(); err != nil {
		log.Error("cannot get available releases", "err", err.Error())
		os.Exit(1) 
	}
}
