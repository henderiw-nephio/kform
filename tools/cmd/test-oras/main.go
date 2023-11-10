package main

import (
	"context"
	"fmt"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

func main() {
	// 0. Create a file store
	fs, err := file.New("./build/provider-resourcebackend/schemas")
	if err != nil {
		panic(err)
	}
	defer fs.Close()
	ctx := context.Background()

	// 1. Add files to the file store
	mediaType := "application/kform.schema.file"
	fileNames := []string{"provider/resourcebackend.provider.kform.io_providerconfigs.yaml"}
	fileDescriptors := make([]v1.Descriptor, 0, len(fileNames))
	for _, name := range fileNames {
		fileDescriptor, err := fs.Add(ctx, name, mediaType, "")
		if err != nil {
			panic(err)
		}
		fileDescriptors = append(fileDescriptors, fileDescriptor)
		fmt.Printf("file descriptor for %s: %v\n", name, fileDescriptor)
	}

	// 2. Pack the files and tag the packed manifest
	artifactType := "application/kform.schema.artifact"
	opts := oras.PackManifestOptions{
		Layers: fileDescriptors,
	}
	manifestDescriptor, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1_RC4, artifactType, opts)
	if err != nil {
		panic(err)
	}
	fmt.Println("manifest descriptor:", manifestDescriptor)

	tag := "latest"
	if err = fs.Tag(ctx, manifestDescriptor, tag); err != nil {
		panic(err)
	}
}
