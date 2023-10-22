package main

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

const (
	configDir = "./examples"
	fileName  = "testfile.tgz"
	outDir    = "./examples/out"
)

func main() {

	img, err := tarball.ImageFromPath(filepath.Join(configDir, fileName), nil)
	if err != nil {
		panic(err)
	}
	r := mutate.Extract(img)
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Printf("Contents of %s: size: %d\n", hdr.Name, hdr.Size)

		buf := new(bytes.Buffer)
		buf.ReadFrom(tr)
		s := buf.String()
		fmt.Println(s)
		fmt.Println()

		path := filepath.Join(outDir, hdr.Name)
		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		n, err := f.Write(buf.Bytes())
		if err != nil {
			panic(err)
		}
		fmt.Printf("wrote %d bytes of %d\n", n, hdr.Size)

	}

}
