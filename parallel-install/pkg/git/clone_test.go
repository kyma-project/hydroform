package git

import (
	"os"
	"path"
	"testing"

	"github.com/alcortesm/tgz"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type fakeCloner struct {
	repo *git.Repository
}

func (fc *fakeCloner) Clone(url, path string, noCheckout bool) (*git.Repository, error) {
	return fc.repo, nil
}

// TestCloneRepo tests CloneRepo function that is provided with a dummy git repository (no actual cloning is performed)
// The repo has following commits
// 1. Add README (tagged with 1.0.0)
// 2. Update README (tagged with 2.0.0 - HEAD)
func TestCloneRepo(t *testing.T) {
	localRepoRootPath, err := tgz.Extract("testdata/repo.tgz")
	defer func() {
		require.NoError(t, os.RemoveAll(localRepoRootPath))
	}()
	require.NoError(t, err)
	require.NotEmpty(t, localRepoRootPath)

	repo, err := git.PlainOpen(path.Join(localRepoRootPath, "repo"))
	require.NoError(t, err)

	var refs []*plumbing.Reference
	iter, err := repo.References()
	require.NoError(t, err)
	iter.ForEach(func(r *plumbing.Reference) error {
		refs = append(refs, r)
		return nil
	})

	defaultCloner = &fakeCloner{repo: repo}
	defaultLister = &fakeRefLister{refs: refs}

	headRef, err := repo.Head()
	require.NoError(t, err)

	commit, err := repo.CommitObject(headRef.Hash())
	require.NoError(t, err)
	require.Equal(t, "Update README\n", commit.Message)

	err = CloneRepo("github.com/foo", "bar/baz", "1.0.0")
	require.NoError(t, err)

	headRef, err = repo.Head()
	require.NoError(t, err)

	commit, err = repo.CommitObject(headRef.Hash())
	require.NoError(t, err)
	require.Equal(t, "Add README\n", commit.Message)
}
