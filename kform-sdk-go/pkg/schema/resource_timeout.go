package schema

import "time"

type ResourceTimeout struct {
	Create, Read, Default *time.Duration
}
