package cache

import "fmt"

type NSN struct {
	Namespace string
	Name      string
}

func (r NSN) String() string {
	return fmt.Sprintf("%s.%s", r.Namespace, r.Name)
}
