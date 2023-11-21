package oras

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	credentials "github.com/oras-project/oras-credentials-go"
	"github.com/pkg/errors"

	//"oras.land/oras-go/pkg/auth"
	//dockerauth "oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"

	//"oras.land/oras-go/v2/oras"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	PackageMetaLayerMediaType  = "application/vnd.cncf.kform.package.meta.v1.tar+gzip"
	PackageImageLayerMediaType = "application/vnd.cncf.kform.package.image.v1.tar+gzip"
	ModuleMediaType            = "application/vnd.cncf.kform.module.v1+json"
	ProviderMediaType          = "application/vnd.cncf.kform.provider.v1+json"
)

func EmptyCredential(ctx context.Context, hostport string) (auth.Credential, error) {
	return auth.EmptyCredential, nil
}

func DefaultCredential(registry string) auth.CredentialFunc {
	store, err := credentials.NewStoreFromDocker(credentials.StoreOptions{})
	if err != nil {
		return EmptyCredential
	}
	return func(ctx context.Context, registry string) (auth.Credential, error) {
		registry = credentials.ServerAddressFromHostname(registry)
		if registry == "" {
			return auth.EmptyCredential, nil
		}
		return store.Get(ctx, registry)
	}
}

/*
func DefaultCredFunc(ctx context.Context, reg string) (auth.Credential, error) {
	dockerClient, ok := client.authorizer.(*dockerauth.Client)
	if !ok {
		return auth.EmptyCredential, errors.New("unable to obtain docker client")
	}

	username, password, err := dockerClient.Credential(reg)
	if err != nil {
		return auth.EmptyCredential, errors.New("unable to retrieve credentials")
	}

	// A blank returned username and password value is a bearer token
	if username == "" && password != "" {
		return auth.Credential{
			RefreshToken: password,
		}, nil
	}

	return auth.Credential{
		Username: username,
		Password: password,
	}, nil
}
*/

/*
type Client struct {
	debug       bool
	enableCache bool
	out         io.Writer
	authorizer  auth.Client
	//registryAuthorizer *auth.Client
	resolver   func(ref registry.Reference) (remotes.Resolver, error)
	httpClient *http.Client
	plainHTTP  bool
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
		authClient, err := auth.NewClientWithDockerFallback()
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
	var cache auth.Cache
	if client.enableCache {
		cache = auth.DefaultCache
	}
	if client.authorizer == nil {
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

func ClientOptAuth(authorizer auth.Client) ClientOption {
	return func(client *Client) {
		client.authorizer = authorizer
	}
}

type Result struct {
	Manifest *descriptorSummary `json:"manifest"`
	Config   *descriptorSummary `json:"config"`
	PkgMeta  *descriptorSummary `json:"pkgMeta"`
	Image    *descriptorSummary `json:"image"`
	Ref      string             `json:"ref"`
}

type descriptorSummary struct {
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
	Data   []byte `json:"-"`
}
*/

func Push(ctx context.Context, kind v1alpha1.PkgKind, ref string, pkgData []byte, imgData []byte, credfunc auth.CredentialFunc) error {
	// parse the reference
	parsedRef, err := registry.ParseReference(ref)
	if err != nil {
		return errors.Wrap(err, "cannot parse reference")
	}
	// src -> memory
	src := memory.New()
	// dst -> registry
	reg, err := remote.NewRegistry(parsedRef.Registry)
	if err != nil {
		return errors.Wrap(err, "cannot create registry")
	}
	if credfunc == nil {
		credfunc = DefaultCredential(parsedRef.Registry)
	}
	reg.Client = &auth.Client{
		Credential: credfunc,
		Header: http.Header{
			"User-Agent": {"kform"},
		},
	}

	dst, err := reg.Repository(ctx, parsedRef.Repository)
	if err != nil {
		return errors.Wrap(err, "cannot get repository")
	}
	// collect layer descriptors and artifact type based on package type (provider/module)
	layerDescriptors := []ocispecv1.Descriptor{}
	artifactType := ModuleMediaType
	pkgMetaDescriptor, err := pushBlob(ctx, PackageMetaLayerMediaType, pkgData, src)
	if err != nil {
		return err
	}
	layerDescriptors = append(layerDescriptors, pkgMetaDescriptor)
	if kind == v1alpha1.PkgKindProvider {
		artifactType = ProviderMediaType
		imageDescriptor, err := pushBlob(ctx, PackageImageLayerMediaType, imgData, src)
		if err != nil {
			return err
		}
		layerDescriptors = append(layerDescriptors, imageDescriptor)

	}
	// generate manifest and push from src (memory store) to dst (remote registry)
	manifestDesc, err := oras.PackManifest(
		ctx,
		src,
		oras.PackManifestVersion1_1_RC4,
		artifactType,
		oras.PackManifestOptions{
			Layers:              layerDescriptors,
			ManifestAnnotations: map[string]string{},
		},
	)
	if err != nil {
		return err
	}
	err = src.Tag(ctx, manifestDesc, parsedRef.Reference)
	if err != nil {
		panic(err)
	}

	if _, err := oras.Copy(ctx, src, parsedRef.Reference, dst, "", oras.DefaultCopyOptions); err != nil {
		return err
	}
	//fmt.Fprintf(c.out, "Pushed: %s\n", parsedRef.String())
	//fmt.Fprintf(c.out, "Digest: %s\n", desc.Digest)
	return nil
}

func pushBlob(ctx context.Context, mediaType string, blob []byte, target oras.Target) (ocispecv1.Descriptor, error) {
	desc := content.NewDescriptorFromBytes(mediaType, blob)
	return desc, target.Push(ctx, desc, bytes.NewReader(blob)) // Push the blob to the registry target
}

/*
func (c *Client) Push(kind v1alpha1.PkgKind, ref string, pkgData []byte, imgData []byte) (*Result, error) {
	parsedRef, err := registry.ParseReference(ref)
	if err != nil {
		return nil, err
	}

	memoryStore := memory.New()

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

		//	Image: &descriptorPushSummary{
		//		Digest: imageDescriptor.Digest.String(),
		//		Size:   imageDescriptor.Size,
		//	},

		Ref: parsedRef.String(),
	}
	//fmt.Fprintf(c.out, "Pushed: %s\n", result.Ref)
	//fmt.Fprintf(c.out, "Digest: %s\n", result.Manifest.Digest)
	return result, nil
}
*/

func Pull(ctx context.Context, ref string, credfunc auth.CredentialFunc) error {
	parsedRef, err := registry.ParseReference(ref)
	if err != nil {
		return err
	}
	fmt.Println("ref", parsedRef.String())
	// dst -> memory
	dst := memory.New()
	// dst -> registry
	fmt.Println("registry", parsedRef.Registry)
	reg, err := remote.NewRegistry(parsedRef.Registry)
	if err != nil {
		return errors.Wrap(err, "cannot get registry")
	}
	if credfunc == nil {
		credfunc = DefaultCredential(parsedRef.Registry)
	}
	reg.Client = &auth.Client{
		Credential: credfunc,
		Header: http.Header{
			"User-Agent": {"kform"},
		},
	}
	if err := reg.Ping(ctx); err != nil {
		return errors.Wrap(err, "registry v2 not implemented")
	}
	src, err := reg.Repository(ctx, parsedRef.Repository)
	if err != nil {
		return errors.Wrap(err, "cannot get repository")
	}
	desc, err := oras.Copy(ctx, src, parsedRef.String(), dst, "", oras.DefaultCopyOptions)
	if err != nil {
		return errors.Wrap(err, "cannot copy")
	}

	fmt.Println(desc)
	return nil
}

//https://ghcr.io/kformdev/provider-resourcebackend/provider-resourcebackend/

/*
func (c *Client) Pull(ctx context.Context, ref string) (*Result, error) {
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


	//	if _, err := oci.ReadTgz(result.Schemas.Data); err != nil {
	//		return nil, err
	//	}

	return result, nil
}
*/
