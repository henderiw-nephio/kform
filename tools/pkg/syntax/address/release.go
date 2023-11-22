package address

import (
	"context"
	"strings"

	"github.com/henderiw/logger/log"
)

func (r Releases) GetRelease(version string) *Release {
	for _, release := range r {
		if strings.Contains(release.TagName, version) {
			return &release
		}
	}
	return nil
}

type Images []Image

type Image struct {
	Name    string
	Version string
	Platform
	URL string
}

func (r *Release) GetImageData(ctx context.Context) Images {
	log := log.FromContext(ctx)
	images := Images{}
	for _, asset := range r.Assets {
		if asset.ContentType == "application/octet-stream" && asset.State == "uploaded" {
			split := strings.Split(asset.Name, "_")
			if len(split) != 4 {
				log.Error("wrong release name: expecting <name>_<version>_<os>_<arch>", "got", asset.Name)
				continue
			}
			images = append(images, Image{
				Name:    split[0],
				Version: split[1],
				Platform: Platform{
					OS:   split[2],
					Arch: split[3],
				},
				URL: asset.BrowserDownloadURL,
			})
		}
	}
	return images
}

//"content_type": "application/octet-stream",
// "state": "uploaded",