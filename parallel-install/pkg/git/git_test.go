package git

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-git-fixtures.v3"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

const (
	repo       = "https://github.com/kyma-project/kyma"
	kyma117Rev = "2292dd21453af5f4f517c1c42a1cf5413d8c461c"
)

type fakeLister struct {
}

func (fl *fakeLister) List(repoURL string) ([]*plumbing.Reference, error) {
	return []*plumbing.Reference{
		plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), plumbing.ZeroHash),
		plumbing.NewHashReference(plumbing.NewTagReferenceName("1.15.0"), plumbing.ZeroHash),
		plumbing.NewHashReference(plumbing.ReferenceName("refs/pull/9999/head"), plumbing.ZeroHash),
	}, nil
}

type fakeCloner struct {
}

func (fc *fakeCloner) Clone(repoURL, dstPath string, noCheckout bool) (*git.Repository, error) {
	fixtures.Init()
	f := fixtures.Basic().ByTag("worktree").One()
	return git.PlainOpen(f.Worktree().Root())
}

func TestCloneRevision(t *testing.T) {
	t.Parallel()

	c := client{
		cloner: &fakeCloner{},
		lister: &fakeLister{},
	}

	os.RemoveAll("./clone") //ensure clone folder does not exist
	err := c.CloneRevision(repo, "./clone", "d2e42ddd68eacbb6034e7724e0dd4117ff1f01ee")
	defer os.RemoveAll("./clone")

	require.NoError(t, err, "Cloning Kyma 1.17 should not error")
	_, err = os.Stat("./clone")
	require.NoError(t, err, "cloned local kyma folder should not error")
	_, err = os.Stat("./clone/resources")
	require.NoError(t, err, "cloned local charts folder should not error")
}

// TestResolveRevision tests implicitly also the commit ID resolution functions for: Branch, PR and Tag
func TestResolveRevision(t *testing.T) {
	t.Parallel()

	c := client{
		lister: &fakeLister{},
	}

	// main branch head
	r, err := c.ResolveRevision(repo, "main")
	require.NoError(t, err, "Resolving Kyma's main revision should not error")
	require.True(t, isHex(r), "The resolved main revision should be a hex string")
	// version tag
	r, err = c.ResolveRevision(repo, "1.15.0")
	require.NoError(t, err, "Resolving Kyma's 1.15.0 version tag should not error")
	require.True(t, isHex(r), "The resolved 1.15.0 version tag revision should be a hex string")

	// Pull Request
	r, err = c.ResolveRevision(repo, "PR-9999")
	require.NoError(t, err, "Resolving Kyma's Pull request head should not error")
	require.True(t, isHex(r), "The resolved Pull request head should be a hex string")

	// Bad ref
	_, err = c.ResolveRevision(repo, "not-a-git-ref")
	require.Error(t, err)
}
