package address

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/apparentlymart/go-versions/versions"
)

/*
https://github.com/henderiw-nephio/kform/releases/download/v0.0.1/provider-kubernetes_0.0.1_darwin_amd64
europe-docker.pkg.dev/srlinux/eu.gcr.io/provider-xxxx
github.com/henderiw-nephio/kform/provider-xxxx
*/

// .kform/providers/github.com/henderiw-nephio_kform/kubernetes/0.0.1/darwin_arm64/<binary>
// .kform/providers/kubernetes/<binary>
// .kform/providers/<hostname>/<namespace>/<provider-name>/<version>/<platform>/<binary>
//

type Address struct {
	HostName  string
	Namespace string
	Name      string
}

func (r Address) IsLocal() bool {
	return r.HostName == "" || r.HostName == "."
}

func (r Address) Path() string {
	return filepath.Join(
		r.HostName,
		strings.Join(strings.Split(r.Namespace, "/"), "_"), // replaces / with _
		r.Name,
	)
}

func (r Address) ProjectName() string {
	split := strings.Split(r.Namespace, "/")
	return split[len(split)-1]
}

type Platform struct {
	OS, Arch string
}

func (r Platform) String() string {
	return fmt.Sprintf("%s_%s", r.OS, r.Arch)
}

type Package struct {
	Type               PackageType
	Address            *Address
	Platform           *Platform
	AvailableVersions  versions.List
	VersionConstraints string
	CandidateVersions  versions.List
	SelectedVersion    string
}

type PackageType string

const (
	PackageTypeProvider PackageType = "provider"
	PackageTypeModule   PackageType = "module"
)

/*
func (r *Package) GetVersion() string {
	if len(r.Version) == 0 {
		return ""
	}
	if r.Version[0] == 'v' {
		return r.Version
	}
	return fmt.Sprintf("v%s", r.Version)
}

func (r *Package) GetRawVersion() string {
	if len(r.Version) == 0 {
		return ""
	}
	if r.Version[0] == 'v' {
		return string(r.Version[1:])
	}
	return r.Version
}
*/

func (r *Package) githubDownloadPath(version string) string {
	return filepath.Join(r.Address.Namespace, "releases", "download", fmt.Sprintf("v%s", version), r.Filename(version))
}

// filename is aligned with go releaser
func (r *Package) Filename(version string) string {
	if r.IsLocal() {
		return fmt.Sprintf("%s-%s", r.Type, r.Address.Name)
	}
	return fmt.Sprintf("%s-%s_%s_%s", r.Type, r.Address.Name, version, r.Platform.String())
}

func (r *Package) githubChecksumPath(version string) string {
	return filepath.Join(r.Address.Namespace, "releases", "download", fmt.Sprintf("v%s", version), r.checksumFilename(version))
}

func (r *Package) githubReleasesPath() string {
	return filepath.Join("repos", r.Address.Namespace, "releases")
}

// filename is aligned with go releaser
func (r *Package) checksumFilename(version string) string {
	return fmt.Sprintf("%s_%s_checksums.txt", r.Address.ProjectName(), version)
}

func (r *Package) URL(version string) string {
	u := url.URL{
		Scheme: "https",
		Host:   r.Address.HostName,
		Path:   r.githubDownloadPath(version),
	}
	return u.String()
}

func (r *Package) ChecksumURL(version string) string {
	u := url.URL{
		Scheme: "https",
		Host:   r.Address.HostName,
		Path:   r.githubChecksumPath(version),
	}
	return u.String()
}

func (r *Package) ReleasesURL() string {
	u := url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   r.githubReleasesPath(),
	}
	return u.String()
}

func (r *Package) BasePath() string {
	return r.Address.Path()
}

func (r *Package) FilePath(version string) string {
	if r.Address.IsLocal() {
		return filepath.Join(r.Address.Path(), r.Filename(version))
	}
	return filepath.Join(r.Address.Path(), version, r.Platform.String(), r.Filename(version))
}

func (r *Package) FilePathWithSelectedVersion() string {
	return r.FilePath(r.SelectedVersion)
}

func (r *Package) DirPath(version string) string {
	if r.Address.IsLocal() {
		return r.Address.Path()
	}
	return filepath.Join(r.Address.Path(), version, r.Platform.String())
}

func (r *Package) IsLocal() bool {
	return r.Address.IsLocal()
}

// ParseSource return a registry hostname and namespace
// if the source is an empty string or . the source is local to the filesystem, returns empty hostname and namespace
// if the source is not local we expect a delineation with a / between a registry hostname and a namespace
func ParseSource(source string) (string, string, error) {
	if source == "" || source == "." {
		// this is a local package -> we dont verify the version, etc etc
		return "", "", nil
	}
	split := strings.Split(source, "/")
	if len(split) < 3 {
		return "", "", fmt.Errorf("a source format has the following format <hostname>/<namespace>, got: %s", source)
	}
	return split[0], filepath.Join(split[1:]...), nil
}

func GetVersionFromPath(path string) string {
	split := strings.Split(path, "/")
	if len(split) >= 6 {
		return split[3]
	}
	return ""
}
