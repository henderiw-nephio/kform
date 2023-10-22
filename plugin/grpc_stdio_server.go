package plugin

import (
	"bufio"
	"io"
	"log/slog"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/henderiw-nephio/kform/plugin/internal/plugin"
)

// grpcStdioBuffer is the buffer size we try to fill when sending a chunk of
// stdio data. This is currently 1 KB for no reason other than that seems like
// enough (stdio data isn't that common) and is fairly low.
const grpcStdioBuffer = 1 * 1024

// grpcStdioServer implements the Stdio service and streams stdiout/stderr.
type grpcStdioServer struct {
	plugin.UnimplementedGRPCStdioServer
	stdoutCh <-chan []byte
	stderrCh <-chan []byte
}

// newGRPCStdioServer creates a new grpcStdioServer and starts the stream
// copying for the given out and err readers.
//
// This must only be called ONCE per srcOut, srcErr.
func newGRPCStdioServer(l *slog.Logger, srcOut, srcErr io.Reader) *grpcStdioServer {
	stdoutCh := make(chan []byte)
	stderrCh := make(chan []byte)

	// Begin copying the streams
	go copyChan(l, stdoutCh, srcOut)
	go copyChan(l, stderrCh, srcErr)

	// Construct our server
	return &grpcStdioServer{
		stdoutCh: stdoutCh,
		stderrCh: stderrCh,
	}
}

// StreamStdio streams our stdout/err as the response.
func (s *grpcStdioServer) StreamStdio(
	_ *empty.Empty,
	srv plugin.GRPCStdio_StreamStdioServer,
) error {
	// Share the same data value between runs. Sending this over the wire
	// marshals it so we can reuse this.
	var data plugin.StdioData

	for {
		// Read our data
		select {
		case data.Data = <-s.stdoutCh:
			data.Channel = plugin.StdioData_STDOUT

		case data.Data = <-s.stderrCh:
			data.Channel = plugin.StdioData_STDERR

		case <-srv.Context().Done():
			return nil
		}

		// Not sure if this is possible, but if we somehow got here and
		// we didn't populate any data at all, then just continue.
		if len(data.Data) == 0 {
			continue
		}

		// Send our data to the client.
		if err := srv.Send(&data); err != nil {
			return err
		}
	}
}

// copyChan copies an io.Reader into a channel.
func copyChan(l *slog.Logger, dst chan<- []byte, src io.Reader) {
	bufsrc := bufio.NewReader(src)

	for {
		// Make our data buffer. We allocate a new one per loop iteration
		// so that we can send it over the channel.
		var data [grpcStdioBuffer]byte

		// Read the data, this will block until data is available
		n, err := bufsrc.Read(data[:])

		// We have to check if we have data BEFORE err != nil. The bufio
		// docs guarantee n == 0 on EOF but its better to be safe here.
		if n > 0 {
			// We have data! Send it on the channel. This will block if there
			// is no reader on the other side. We expect that go-plugin will
			// connect immediately to the stdio server to drain this so we want
			// this block to happen for backpressure.
			dst <- data[:n]
		}

		// If we hit EOF we're done copying
		if err == io.EOF {
			l.Debug("stdio EOF, exiting copy loop")
			return
		}

		// Any other error we just exit the loop. We don't expect there to
		// be errors since our use case for this is reading/writing from
		// a in-process pipe (os.Pipe).
		if err != nil {
			l.Warn("error copying stdio data, stopping copy", "err", err)
			return
		}
	}
}
