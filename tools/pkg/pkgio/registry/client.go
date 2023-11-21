package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/containerd/containerd/remotes"
	"github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/auth"
	dockerauth "oras.land/oras-go/pkg/auth/docker"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
	registryauth "oras.land/oras-go/pkg/registry/remote/auth"
	"oras.land/oras-go/v2/registry"
)

const (
	PackageMetaLayerMediaType  = "application/vnd.cncf.kform.package.meta.v1.tar+gzip"
	PackageImageLayerMediaType = "application/vnd.cncf.kform.package.image.v1.tar+gzip"
	ModuleMediaType            = "application/vnd.cncf.kform.module.v1+json"
	ProviderMediaType          = "application/vnd.cncf.kform.provider.v1+json"
)

type Client struct {
	debug              bool
	enableCache        bool
	out                io.Writer
	authorizer         auth.Client
	registryAuthorizer *registryauth.Client
	resolver           func(ref registry.Reference) (remotes.Resolver, error)
	httpClient         *http.Client
	plainHTTP          bool
}

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) (*Client, error) {
	client := &Client{
		out: io.Discard,
	}

	for _, o := range opts {
		o(client)
	}
	if client.authorizer == nil {
		authClient, err := dockerauth.NewClientWithDockerFallback()
		if err != nil {
			return nil, err
		}
		client.authorizer = authClient
	}
	resolverFn := client.resolver // copy for avoiding recursive call
	client.resolver = func(ref registry.Reference) (remotes.Resolver, error) {
		if resolverFn != nil {
			// validate if the resolverFn returns a valid resolver
			if resolver, err := resolverFn(ref); resolver != nil && err == nil {
				return resolver, nil
			}
		}
		headers := http.Header{
			"User-Agent": {"kform"},
		}
		opts := []auth.ResolverOption{auth.WithResolverHeaders(headers)}
		if client.httpClient != nil {
			opts = append(opts, auth.WithResolverClient(client.httpClient))
		}
		if client.plainHTTP {
			opts = append(opts, auth.WithResolverPlainHTTP())
		}
		resolver, err := client.authorizer.ResolverWithOpts(opts...)
		if err != nil {
			return nil, err
		}
		return resolver, nil
	}

	// allocate a cache if option is set
	var cache registryauth.Cache
	if client.enableCache {
		cache = registryauth.DefaultCache
	}
	if client.registryAuthorizer == nil {
		client.registryAuthorizer = &registryauth.Client{
			Client: client.httpClient,
			Header: http.Header{
				"User-Agent": {"kform"},
			},
			Cache: cache,
			Credential: func(ctx context.Context, reg string) (registryauth.Credential, error) {
				dockerClient, ok := client.authorizer.(*dockerauth.Client)
				if !ok {
					return registryauth.EmptyCredential, errors.New("unable to obtain docker client")
				}

				username, password, err := dockerClient.Credential(reg)
				if err != nil {
					return registryauth.EmptyCredential, errors.New("unable to retrieve credentials")
				}

				// A blank returned username and password value is a bearer token
				if username == "" && password != "" {
					return registryauth.Credential{
						RefreshToken: password,
					}, nil
				}

				return registryauth.Credential{
					Username: username,
					Password: password,
				}, nil
			},
		}
	}
	return client, nil
}

// ClientOptResolver returns a function that sets the resolver setting on a client options set
func ClientOptResolver(resolver remotes.Resolver) ClientOption {
	return func(client *Client) {
		client.resolver = func(ref registry.Reference) (remotes.Resolver, error) {
			return resolver, nil
		}
	}
}

type Result struct {
	Manifest *descriptorSummary `json:"manifest"`
	Config   *descriptorSummary `json:"config"`
	PkgMeta  *descriptorSummary `json:"pkgMeta"`
	Image    *descriptorSummary `json:"image"`
	Ref      string                 `json:"ref"`
}

type descriptorSummary struct {
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
	Data   []byte `json:"-"`
}

func (c *Client) Push(kind v1alpha1.PkgKind, ref string, pkgData []byte, imgData []byte) (*Result, error) {
	parsedRef, err := registry.ParseReference(ref)
	if err != nil {
		return nil, err
	}

	//fmt.Println("schemaData", schemaData)
	memoryStore := content.NewMemory()
	// collect layer descriptors
	descriptors := []ocispecv1.Descriptor{}
	pkgMetaDescriptor, err := memoryStore.Add("pkgMeta", PackageMetaLayerMediaType, pkgData)
	if err != nil {
		return nil, err
	}
	descriptors = append(descriptors, pkgMetaDescriptor)
	if kind == v1alpha1.PkgKindProvider {
		imageDescriptor, err := memoryStore.Add("image", PackageImageLayerMediaType, imgData)
		if err != nil {
			return nil, err
		}
		descriptors = append(descriptors, imageDescriptor)
	}

	// collect config descriptors
	var configDescriptor ocispecv1.Descriptor
	if kind == v1alpha1.PkgKindProvider {
		configDescriptor, err = memoryStore.Add("config", ProviderMediaType, []byte("config"))
		if err != nil {
			return nil, err
		}
	} else {
		// module
		configDescriptor, err = memoryStore.Add("config", ModuleMediaType, []byte("config"))
		if err != nil {
			return nil, err
		}
	}

	ociAnnotations := map[string]string{}

	// generate manifest
	manifestData, manifest, err := content.GenerateManifest(&configDescriptor, ociAnnotations, descriptors...)
	if err != nil {
		return nil, err
	}
	// store manifest in the memory store
	if err := memoryStore.StoreManifest(parsedRef.String(), manifest, manifestData); err != nil {
		return nil, err
	}
	// resolve the remote registry based on the ref
	remotesResolver, err := c.resolver(parsedRef)
	if err != nil {
		return nil, err
	}
	registryStore := content.Registry{Resolver: remotesResolver}
	_, err = oras.Copy(context.Background(), memoryStore, parsedRef.String(), registryStore, "",
		oras.WithNameValidation(nil))
	if err != nil {
		return nil, err
	}
	result := &Result{
		Manifest: &descriptorSummary{
			Digest: manifest.Digest.String(),
			Size:   manifest.Size,
		},
		Config: &descriptorSummary{
			Digest: configDescriptor.Digest.String(),
			Size:   configDescriptor.Size,
		},
		PkgMeta: &descriptorSummary{
			Digest: pkgMetaDescriptor.Digest.String(),
			Size:   pkgMetaDescriptor.Size,
		},
		/*
			Image: &descriptorPushSummary{
				Digest: imageDescriptor.Digest.String(),
				Size:   imageDescriptor.Size,
			},
		*/
		Ref: parsedRef.String(),
	}
	fmt.Fprintf(c.out, "Pushed: %s\n", result.Ref)
	fmt.Fprintf(c.out, "Digest: %s\n", result.Manifest.Digest)
	return result, nil
}

func (c *Client) Pull(ref string) (*Result, error) {
	parsedRef, err := registry.ParseReference(ref)
	if err != nil {
		return nil, err
	}
	memoryStore := content.NewMemory()
	allowedMediaTypes := []string{
		PackageMetaLayerMediaType,
		PackageImageLayerMediaType,
		ProviderMediaType,
		ModuleMediaType,
	}
	minNumDescriptors := 2

	descriptors := []ocispecv1.Descriptor{}
	layers := []ocispecv1.Descriptor{}
	remotesResolver, err := c.resolver(parsedRef)
	if err != nil {
		return nil, err
	}
	registryStore := content.Registry{Resolver: remotesResolver}

	manifest, err := oras.Copy(context.Background(), registryStore, parsedRef.String(), memoryStore, "",
		oras.WithPullEmptyNameAllowed(),
		oras.WithAllowedMediaTypes(allowedMediaTypes),
		oras.WithLayerDescriptors(func(l []ocispecv1.Descriptor) {
			layers = l
		}))
	if err != nil {
		return nil, err
	}

	descriptors = append(descriptors, manifest)
	descriptors = append(descriptors, layers...)

	fmt.Println("descriptor length", len(descriptors))
	if len(descriptors) < minNumDescriptors {
		return nil, fmt.Errorf("manifest does not contain minimum number of descriptors (%d), descriptors found: %d",
			minNumDescriptors, len(descriptors))
	}
	var configDescriptor *ocispecv1.Descriptor
	var pkgMetaDescriptor *ocispecv1.Descriptor
	var imageDescriptor *ocispecv1.Descriptor
	for _, descriptor := range descriptors {
		d := descriptor
		switch d.MediaType {
		case ProviderMediaType, ModuleMediaType:
			configDescriptor = &d
		case PackageMetaLayerMediaType:
			pkgMetaDescriptor = &d
		case PackageImageLayerMediaType:
			imageDescriptor = &d
		case manifest.MediaType:
		default:
			fmt.Println("unexpected descriptor", d.MediaType, d.Digest.String(), d.Size)
			if _, data, ok := memoryStore.Get(d); !ok {
				return nil, fmt.Errorf("unable to retrieve config with digest %s", d.Digest.String())
			} else {
				fmt.Println("unexpected data", string(data))
			}
		}
	}
	fmt.Println("ArtifactType:", manifest.ArtifactType)
	result := &Result{
		Manifest: &descriptorSummary{
			Digest: manifest.Digest.String(),
			Size:   manifest.Size,
		},
		Config: &descriptorSummary{
			Digest: configDescriptor.Digest.String(),
			Size:   configDescriptor.Size,
		},
		PkgMeta: &descriptorSummary{
			Digest: pkgMetaDescriptor.Digest.String(),
			Size:   pkgMetaDescriptor.Size,
		},
		Image: &descriptorSummary{
			Digest: imageDescriptor.Digest.String(),
			Size:   imageDescriptor.Size,
		},
		Ref: parsedRef.String(),
	}
	fmt.Fprintf(c.out, "Pulled: %s\n", result.Ref)
	fmt.Fprintf(c.out, "Digest: %s\n", result.Manifest.Digest)

	if _, manifestData, ok := memoryStore.Get(manifest); !ok {
		return nil, fmt.Errorf("unable to retrieve manifest blob with digest %s", manifest.Digest)
	} else {
		result.Manifest.Data = manifestData
	}
	if _, configData, ok := memoryStore.Get(*configDescriptor); !ok {
		return nil, fmt.Errorf("unable to retrieve config with digest %s", configDescriptor.Digest)
	} else {
		result.Config.Data = configData
	}
	if _, pkgMetaData, ok := memoryStore.Get(*pkgMetaDescriptor); !ok {
		return nil, fmt.Errorf("unable to retrieve pkgMetaData with digest %s", pkgMetaDescriptor.Digest)
	} else {
		result.PkgMeta.Data = pkgMetaData
	}
	if _, imageData, ok := memoryStore.Get(*imageDescriptor); !ok {
		return nil, fmt.Errorf("unable to retrieve image with digest %s", imageDescriptor.Digest)
	} else {
		result.Image.Data = imageData
	}

	fmt.Println("ref", result.Ref)
	fmt.Println("manifest", string(result.Manifest.Data))
	fmt.Println("config", string(result.Config.Data))
	//fmt.Println("schemas", string(result.Schemas.Data))
	fmt.Println("image", string(result.Image.Data))

	/*
		if _, err := oci.ReadTgz(result.Schemas.Data); err != nil {
			return nil, err
		}
	*/
	return result, nil
}
