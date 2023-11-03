package client

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	client.Reader
	client.Writer
	Applicator
}

// An ApplyFn is a function that satisfies the Applicator interface.
type ApplyFn func(context.Context, client.Object, ...ApplyOption) error

// Apply changes to the supplied object.
func (fn ApplyFn) Apply(ctx context.Context, o client.Object, ao ...ApplyOption) error {
	return fn(ctx, o, ao...)
}

// An Applicator applies changes to an object.
type Applicator interface {
	Apply(context.Context, client.Object, ...ApplyOption) error
}

// An ApplyOption is called before patching the current object to match the
// desired object. ApplyOptions are not called if no current object exists.
type ApplyOption func(ctx context.Context, current, desired runtime.Object) error
