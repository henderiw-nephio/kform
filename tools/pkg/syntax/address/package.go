package address

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/apparentlymart/go-versions/versions"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/pkg/errors"
)

/*
https://github.com/henderiw-nephio/kform/releases/download/v0.0.1/provider-kubernetes_0.0.1_darwin_amd64
europe-docker.pkg.dev/srlinux/eu.gcr.io/provider-xxxx
github.com/henderiw-nephio/kform/provider-xxxx
*/

// address -> hostname, namespace, name
func GetPackage(nsn cache.NSN, source string) (*Package, error) {
	// TODO handle multiple requirements
	hostname, namespace, err := ParseSource(source)
	if err != nil {
		return nil, err
	}
	pkg := &Package{
		Type: PackageTypeProvider,
		Address: &Address{
			HostName:  hostname,
			Namespace: namespace,
			Name:      nsn.Name,
		},
		Platform: &Platform{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
		VersionConstraints: "",
	}
	return pkg, nil
}

// GetReleases returns the avilable releases/versions of the package
func (r *Package) GetReleases() error {
	url := r.ReleasesURL()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get versions url %s, status code: %d", url, resp.StatusCode)
	}

	availableReleases := []Release{}
	if err := json.NewDecoder(resp.Body).Decode(&availableReleases); err != nil {
		return err
	}
	for _, availableRelease := range availableReleases {
		v, err := versions.ParseVersion(strings.ReplaceAll(availableRelease.TagName, "v", ""))
		if err != nil {
			return fmt.Errorf("cannot parse version: %s, err %s", availableRelease.TagName, err.Error())
		}
		r.AvailableVersions = append(r.AvailableVersions, v)
	}
	fmt.Println("available versions", r.AvailableVersions)

	return nil
}

type Release struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
}

func (r *Package) AddConstraints(constraint string) {
	if r.VersionConstraints == "" {
		r.VersionConstraints = constraint
	} else {
		r.VersionConstraints = fmt.Sprintf("%s, %s", r.VersionConstraints, constraint)
	}
	fmt.Println("constraints", r.VersionConstraints)
}

func (r *Package) GenerateCandidates() error {
	allowed, err := versions.MeetingConstraintsStringRuby(r.VersionConstraints)
	if err != nil {
		return errors.Wrap(err, "invalid version constraint")
	}
	fmt.Println("allowed versions", allowed)
	fmt.Println("available versions", r.AvailableVersions)
	r.CandidateVersions = r.AvailableVersions.Filter(allowed)
	fmt.Println("candidate versions", r.CandidateVersions)
	return nil
}

func (r *Package) GetRemoteChecksum(version string) (string, error) {
	resp, err := http.Get(r.ChecksumURL(version))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download checksum file %s, status code: %d", r.ChecksumURL(version), resp.StatusCode)
	}

	s := bufio.NewScanner(resp.Body)
	for s.Scan() {
		line := s.Text()
		if strings.HasSuffix(line, r.Filename(version)) {
			return strings.TrimSpace(strings.TrimSuffix(line, r.Filename(version))), nil
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("checksum for %s not found in the url %s", r.BasePath(), r.ChecksumURL(version))
}

func (r *Package) HasVersion(version string) bool {
	return r.CandidateVersions.Set().Has(versions.MustParseVersion(version))
}

func (r *Package) Newest() string {
	fmt.Println("candidiate versions", r.CandidateVersions)
	return r.CandidateVersions.Newest().String()
}

func (r *Package) UpdateSelectedVersion(version string) {
	r.SelectedVersion = version
}

func (r *Package) GetSelectedVersion() string {
	return r.SelectedVersion
}