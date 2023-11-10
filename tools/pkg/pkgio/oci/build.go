package oci

import (
	"archive/tar"
	"bytes"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func Build(files map[string]string) (v1.Image, error) {
	// copy files to tarbal
	tarBuf := new(bytes.Buffer)
	tw := tar.NewWriter(tarBuf)
	for fileName, data := range files {
		buf := bytes.NewBufferString(data)
		hdr := &tar.Header{
			Name: fileName,
			Mode: int64(0o644),
			Size: int64(buf.Len()),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := io.Copy(tw, buf); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}

	// Build image layer from tarball.
	layer, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(tarBuf.Bytes())), nil
	})
	if err != nil {
		return nil, err
	}
	// Append layer to empty image.
	return mutate.AppendLayers(Image, layer)
}
