package plugin

import (
	"errors"
	"sync"

	"github.com/henderiw-nephio/kform/plugin/internal/plugin"
)

// sendErr is used to pass errors back during a send.
type sendErr struct {
	i  *plugin.ConnInfo
	ch chan error
}

// gRPCBrokerServer is used by the plugin to start a stream and to send
// connection information to/from the plugin. Implements GRPCBrokerServer and
// streamer interfaces.
type gRPCBrokerServer struct {
	plugin.UnimplementedGRPCBrokerServer
	// send is used to send connection info to the gRPC stream.
	send chan *sendErr

	// recv is used to receive connection info from the gRPC stream.
	recv chan *plugin.ConnInfo

	// quit closes down the stream.
	quit chan struct{}

	// o is used to ensure we close the quit channel only once.
	o sync.Once
}

func newGRPCBrokerServer() *gRPCBrokerServer {
	return &gRPCBrokerServer{
		send: make(chan *sendErr),
		recv: make(chan *plugin.ConnInfo),
		quit: make(chan struct{}),
	}
}

// StartStream implements the GRPCBrokerServer interface and will block until
// the quit channel is closed or the context reports Done. The stream will pass
// connection information to/from the client.
func (s *gRPCBrokerServer) StartStream(stream plugin.GRPCBroker_StartStreamServer) error {
	doneCh := stream.Context().Done()
	defer s.Close()

	// Proccess send stream
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

	// Process receive stream
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
// to the client.
func (s *gRPCBrokerServer) Send(i *plugin.ConnInfo) error {
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
// sent from the client from the stream to the broker.
func (s *gRPCBrokerServer) Recv() (*plugin.ConnInfo, error) {
	select {
	case <-s.quit:
		return nil, errors.New("broker closed")
	case i := <-s.recv:
		return i, nil
	}
}

// Close closes the quit channel, shutting down the stream.
func (s *gRPCBrokerServer) Close() {
	s.o.Do(func() {
		close(s.quit)
	})
}
