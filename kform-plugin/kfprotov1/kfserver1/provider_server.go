package kfserver1

import (
	"context"
	"log/slog"
	"sync"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1"
	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
	"github.com/henderiw/logger/log"
)

// New converts a kfprotov1.ProviderServer into a server capable of handling
// kform protocol requests and issuing responses using the gRPC types.
func New(name string, serve kfprotov1.ProviderServer, opts ...ServeOpt) kfplugin1.ProviderServer {
	conf := ServeConfig{}
	for _, opt := range opts {
		err := opt.ApplyServeOpt(&conf)
		if err != nil {
			// this should never happen, we already executed all
			// this code as part of Serve
			panic(err)
		}
	}
	return &server{
		name:     name,
		provider: serve,
		stopCh:   make(chan struct{}),
		l:        log.NewLogger(&log.HandlerOptions{Name: "server-proxy", AddSource: false}),
	}
}

type server struct {
	name     string
	provider kfprotov1.ProviderServer
	kfplugin1.UnimplementedProviderServer

	m      sync.Mutex
	stopCh chan struct{}

	l *slog.Logger
}

func mergeStop(ctx context.Context, cancel context.CancelFunc, stopCh chan struct{}) {
	select {
	case <-ctx.Done():
		return
	case <-stopCh:
		cancel()
	}
}

// stoppableContext returns a context that wraps `ctx` but will be canceled
// when the server's stopCh is closed.
//
// This is used to cancel all in-flight contexts when the Stop method of the
// server is called.
func (s *server) cancelContext(ctx context.Context) context.Context {
	s.m.Lock()
	defer s.m.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	go mergeStop(ctx, cancel, s.stopCh)
	return ctx
}

func (s *server) Capabilities(ctx context.Context, in *kfplugin1.Capabilities_Request) (*kfplugin1.Capabilities_Response, error) {
	// todo add ctx + tracing
	rpc := "capabilities"
	ctx = s.cancelContext(ctx)
	log := s.l
	log.Info(rpc)

	resp, err := s.provider.Capabilities(ctx, in)
	if err != nil {
		log.Error(rpc, "error", err)
		return nil, err
	}
	return resp, nil
}

func (s *server) Configure(ctx context.Context, in *kfplugin1.Configure_Request) (*kfplugin1.Configure_Response, error) {
	// todo add ctx + tracing
	rpc := "configure"
	ctx = s.cancelContext(ctx)
	log := s.l
	log.Info(rpc)

	resp, err := s.provider.Configure(ctx, in)
	if err != nil {
		log.Error(rpc, "error", err)
		return nil, err
	}
	return resp, nil
}
func (s *server) StopProvider(ctx context.Context, in *kfplugin1.StopProvider_Request) (*kfplugin1.StopProvider_Response, error) {
	// todo add ctx + tracing
	rpc := "stopProvider"
	ctx = s.cancelContext(ctx)
	log := s.l
	log.Info(rpc)

	resp, err := s.provider.StopProvider(ctx, in)
	if err != nil {
		log.Error(rpc, "error", err)
		return nil, err
	}
	s.stop()
	return resp, nil
}

func (s *server) stop() {
	close(s.stopCh)
	s.stopCh = make(chan struct{})
}

func (s *server) ReadDataSource(ctx context.Context, in *kfplugin1.ReadDataSource_Request) (*kfplugin1.ReadDataSource_Response, error) {
	// todo add ctx + tracing
	rpc := "readDataSource"
	ctx = s.cancelContext(ctx)
	log := s.l
	log.Info(rpc)

	resp, err := s.provider.ReadDataSource(ctx, in)
	if err != nil {
		log.Error(rpc, "error", err)
		return nil, err
	}
	return resp, nil
}

func (s *server) ListDataSource(ctx context.Context, in *kfplugin1.ListDataSource_Request) (*kfplugin1.ListDataSource_Response, error) {
	// todo add ctx + tracing
	rpc := "listDataSource"
	ctx = s.cancelContext(ctx)
	log := s.l
	log.Info(rpc)

	resp, err := s.provider.ListDataSource(ctx, in)
	if err != nil {
		log.Error(rpc, "error", err)
		return nil, err
	}
	return resp, nil
}

func (s *server) ReadResource(ctx context.Context, in *kfplugin1.ReadResource_Request) (*kfplugin1.ReadResource_Response, error) {
	// todo add ctx + tracing
	rpc := "readDataSource"
	ctx = s.cancelContext(ctx)
	log := s.l
	log.Info(rpc)

	resp, err := s.provider.ReadResource(ctx, in)
	if err != nil {
		log.Error(rpc, "error", err)
		return nil, err
	}
	return resp, nil
}

func (s *server) CreateResource(ctx context.Context, in *kfplugin1.CreateResource_Request) (*kfplugin1.CreateResource_Response, error) {
	// todo add ctx + tracing
	rpc := "createDataSource"
	ctx = s.cancelContext(ctx)
	log := s.l
	log.Info(rpc)

	resp, err := s.provider.CreateResource(ctx, in)
	if err != nil {
		log.Error(rpc, "error", err)
		return nil, err
	}
	return resp, nil
}

func (s *server) UpdateResource(ctx context.Context, in *kfplugin1.UpdateResource_Request) (*kfplugin1.UpdateResource_Response, error) {
	// todo add ctx + tracing
	rpc := "updateDataSource"
	ctx = s.cancelContext(ctx)
	log := s.l
	log.Info(rpc)

	resp, err := s.provider.UpdateResource(ctx, in)
	if err != nil {
		log.Error(rpc, "error", err)
		return nil, err
	}
	return resp, nil
}

func (s *server) DeleteResource(ctx context.Context, in *kfplugin1.DeleteResource_Request) (*kfplugin1.DeleteResource_Response, error) {
	// todo add ctx + tracing
	rpc := "deleteDataSource"
	ctx = s.cancelContext(ctx)
	log := s.l
	log.Info(rpc)

	resp, err := s.provider.DeleteResource(ctx, in)
	if err != nil {
		log.Error(rpc, "error", err)
		return nil, err
	}
	return resp, nil
}
