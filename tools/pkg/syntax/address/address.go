package address

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
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
	Type     PackageType
	Address  *Address
	Version  string // semantic versioning
	Platform *Platform
}

type PackageType string

const (
	PackageTypeProvider PackageType = "provider"
	PackageTypeModule   PackageType = "module"
)

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

func (r *Package) githubDownloadPath() string {
	return filepath.Join(r.Address.Namespace, "releases", "download", r.GetVersion(), r.Filename())
}

// filename is aligned with go releaser
func (r *Package) Filename() string {
	if r.IsLocal() {
		return fmt.Sprintf("%s-%s", r.Type, r.Address.Name)
	}
	return fmt.Sprintf("%s-%s_%s_%s", r.Type, r.Address.Name, r.GetRawVersion(), r.Platform.String())
}

func (r *Package) githubChecksumPath() string {
	return filepath.Join(r.Address.Namespace, "releases", "download", r.GetVersion(), r.checksumFilename())
}

// filename is aligned with go releaser
func (r *Package) checksumFilename() string {
	return fmt.Sprintf("%s_%s_checksums.txt", r.Address.ProjectName(), r.GetRawVersion())
}

func (r *Package) URL() string {
	u := url.URL{
		Scheme: "https",
		Host:   r.Address.HostName,
		Path:   r.githubDownloadPath(),
	}
	return u.String()
}

func (r *Package) ChecksumURL() string {
	u := url.URL{
		Scheme: "https",
		Host:   r.Address.HostName,
		Path:   r.githubChecksumPath(),
	}
	return u.String()
}

func (r *Package) Path() string {
	if r.Address.IsLocal() {
		return filepath.Join(r.Address.Path(), r.Filename())
	}
	return filepath.Join(r.Address.Path(), r.Version, r.Platform.String(), r.Filename())
}

func (r *Package) DirPath() string {
	if r.Address.IsLocal() {
		return r.Address.Path()
	}
	return filepath.Join(r.Address.Path(), r.Version, r.Platform.String())
}

func (r *Package) IsLocal() bool {
	return r.Address.IsLocal()
}

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
