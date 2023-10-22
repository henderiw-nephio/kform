package syntax

import "context"

func (r *parser) Validate(ctx context.Context) {
	r.processConfigs(ctx)
}
