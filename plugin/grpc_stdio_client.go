package plugin

import (
	"bytes"
	"context"
	"io"
	"log/slog"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/henderiw-nephio/kform/plugin/internal/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcStdioClient wraps the stdio service as a client to copy
// the stdio data to output writers.
type grpcStdioClient struct {
	log         *slog.Logger
	stdioClient plugin.GRPCStdio_StreamStdioClient
}

// newGRPCStdioClient creates a grpcStdioClient. This will perform the
// initial connection to the stdio service. If the stdio service is unavailable
// then this will be a no-op. This allows this to work without error for
// plugins that don't support this.
func newGRPCStdioClient(
	ctx context.Context,
	log *slog.Logger,
	conn *grpc.ClientConn,
) (*grpcStdioClient, error) {
	client := plugin.NewGRPCStdioClient(conn)

	// Connect immediately to the endpoint
	stdioClient, err := client.StreamStdio(ctx, &empty.Empty{})

	// If we get an Unavailable or Unimplemented error, this means that the plugin isn't
	// updated and linking to the latest version of go-plugin that supports
	// this. We fall back to the previous behavior of just not syncing anything.
	if status.Code(err) == codes.Unavailable || status.Code(err) == codes.Unimplemented {
		log.Warn("stdio service not available, stdout/stderr syncing unavailable")
		stdioClient = nil
		err = nil
	}
	if err != nil {
		return nil, err
	}

	return &grpcStdioClient{
		log:         log,
		stdioClient: stdioClient,
	}, nil
}

// Run starts the loop that receives stdio data and writes it to the given
// writers. This blocks and should be run in a goroutine.
func (c *grpcStdioClient) Run(stdout, stderr io.Writer) {
	// This will be nil if stdio is not supported by the plugin
	if c.stdioClient == nil {
		c.log.Warn("stdio service unavailable, run will do nothing")
		return
	}

	for {
		c.log.Debug("waiting for stdio data")
		data, err := c.stdioClient.Recv()
		if err != nil {
			if err == io.EOF ||
				status.Code(err) == codes.Unavailable ||
				status.Code(err) == codes.Canceled ||
				status.Code(err) == codes.Unimplemented ||
				err == context.Canceled {
				c.log.Debug("received EOF, stopping recv loop", "err", err)
				return
			}

			c.log.Error("error receiving data", "err", err)
			return
		}

		// Determine our output writer based on channel
		var w io.Writer
		switch data.Channel {
		case plugin.StdioData_STDOUT:
			w = stdout

		case plugin.StdioData_STDERR:
			w = stderr

		default:
			c.log.Warn("unknown channel, dropping", "channel", data.Channel)
			continue
		}

		// Write! In the event of an error we just continue.
		c.log.Debug("received data", "channel", data.Channel.String(), "len", len(data.Data))

		if _, err := io.Copy(w, bytes.NewReader(data.Data)); err != nil {
			c.log.Error("failed to copy all bytes", "err", err)
		}
	}
}
