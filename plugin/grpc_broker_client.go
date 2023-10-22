package plugin

import (
	"context"
	"errors"
	"sync"

	"github.com/henderiw-nephio/kform/plugin/internal/plugin"
	"google.golang.org/grpc"
)

// gRPCBrokerClient is used by the client to start a stream and to send
// connection information to/from the client. Implements GRPCBrokerClient and
// streamer interfaces.
type gRPCBrokerClient struct {
	// client is the underlying GRPC client used to make calls to the server.
	client plugin.GRPCBrokerClient

	// send is used to send connection info to the gRPC stream.
	send chan *sendErr

	// recv is used to receive connection info from the gRPC stream.
	recv chan *plugin.ConnInfo

	// quit closes down the stream.
	quit chan struct{}

	// o is used to ensure we close the quit channel only once.
	o sync.Once
}

func newGRPCBrokerClient(conn *grpc.ClientConn) *gRPCBrokerClient {
	return &gRPCBrokerClient{
		client: plugin.NewGRPCBrokerClient(conn),
		send:   make(chan *sendErr),
		recv:   make(chan *plugin.ConnInfo),
		quit:   make(chan struct{}),
	}
}

// StartStream implements the GRPCBrokerClient interface and will block until
// the quit channel is closed or the context reports Done. The stream will pass
// connection information to/from the plugin.
func (s *gRPCBrokerClient) StartStream() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	defer s.Close()

	stream, err := s.client.StartStream(ctx)
	if err != nil {
		return err
	}
	doneCh := stream.Context().Done()

	go func() {
		for {
			select {
			case <-doneCh:
				return
			case <-s.quit:
				return
			case se := <-s.send:
				err := stream.Send(se.i)
				se.ch <- err
			}
		}
	}()

	for {
		i, err := stream.Recv()
		if err != nil {
			return err
		}
		select {
		case <-doneCh:
			return nil
		case <-s.quit:
			return nil
		case s.recv <- i:
		}
	}
}

// Send is used by the GRPCBroker to pass connection information into the stream
// to the plugin.
func (s *gRPCBrokerClient) Send(i *plugin.ConnInfo) error {
	ch := make(chan error)
	defer close(ch)

	select {
	case <-s.quit:
		return errors.New("broker closed")
	case s.send <- &sendErr{
		i:  i,
		ch: ch,
	}:
	}

	return <-ch
}

// Recv is used by the GRPCBroker to pass connection information that has been
// sent from the plugin to the broker.
func (s *gRPCBrokerClient) Recv() (*plugin.ConnInfo, error) {
	select {
	case <-s.quit:
		return nil, errors.New("broker closed")
	case i := <-s.recv:
		return i, nil
	}
}

// Close closes the quit channel, shutting down the stream.
func (s *gRPCBrokerClient) Close() {
	s.o.Do(func() {
		close(s.quit)
	})
}
