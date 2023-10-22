package genall

import (
	"context"
)

type Generator interface {
	// RegisterMarkers registers all markers needed by this Generator
	// into the given registry.
	RegisterMarkers(ctx context.Context) error
	// Generate generates artifacts produced by this marker.
	Generate(ctx context.Context) error
}
