package testutil

import "testing"

// TestGitRepo manages a local git repository for testing
type TestGitRepo struct {
	T *testing.T

	// RepoDirectory is the temp directory of the git repo
	RepoDirectory string

	// DatasetDirectory is the directory of the testdata files
	DatasetDirectory string

	// RepoName is the name of the repository
	RepoName string

	// Commits keeps track of the commit shas for the changes
	// to the repo.
	Commits []string
}

var EmptyReposInfo = &ReposInfo{}

type ReposInfo struct {
	repos map[string]*TestGitRepo
}

func (ri *ReposInfo) ResolveRepoRef(repoRef string) (string, bool) {
	repo, found := ri.repos[repoRef]
	if !found {
		return "", false
	}
	return repo.RepoDirectory, true
}

func (ri *ReposInfo) ResolveCommitIndex(repoRef string, index int) (string, bool) {
	repo, found := ri.repos[repoRef]
	if !found {
		return "", false
	}
	commits := repo.Commits
	if len(commits) <= index {
		return "", false
	}
	return commits[index], true
}
