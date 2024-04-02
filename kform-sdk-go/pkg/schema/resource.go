package schema

import (
	"context"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
)

type Resource struct {
	//Schema     *apiext.JSONSchemaProps
	//SchemaFunc func() *apiext.JSONSchemaProps

	CreateContext CreateContextFunc
	UpdateContext UpdateContextFunc
	DeleteContext DeleteContextFunc
	ReadContext   ReadContextFunc
	ListContext   ListContextFunc
	//CreateWithoutTimeout CreateContextFunc
	//ReadWithoutTimeout   ReadContextFunc

	Timeouts *ResourceTimeout
}

type CreateContextFunc func(context.Context, *ResourceObject, interface{}) ([]byte, diag.Diagnostics)

type UpdateContextFunc func(context.Context, *ResourceObject, interface{}) ([]byte, diag.Diagnostics)

type DeleteContextFunc func(context.Context, *ResourceObject, interface{}) diag.Diagnostics

type ReadContextFunc func(context.Context, *ResourceObject, interface{}) ([]byte, diag.Diagnostics)

type ListContextFunc func(context.Context, *ResourceObject, interface{}) ([]byte, diag.Diagnostics)
