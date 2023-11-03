package k8sclient

import (
	provclient "github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	RESTCOnfig        *rest.Config
	IgnoreAnnotations []string
	IgnoreLabels      []string
}

func New(cfg Config) (provclient.Client, error) {
	c, err := client.New(cfg.RESTCOnfig, client.Options{
		Scheme: runtime.NewScheme(),
	})
	if err != nil {
		return nil, err
	}

	return NewAPIPatchingApplicator(c), nil
}
