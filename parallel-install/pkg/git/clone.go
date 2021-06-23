package git

import (
	"context"
	"fmt"
	"os"
	"strings"

	// "github.com/go-git/go-git/config"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
)

var defaultCloner repoCloner = &remoteRepoCloner{}

// CloneRepo clones the repository in the given URL to the given dstPath and checks out the given revision.
// revision can be 'main', a release version (e.g. 1.4.1), a commit hash (e.g. 34edf09a) or a PR (e.g. PR-9486).
func CloneRepo(url, dstPath, rev string) error {
	if err := os.RemoveAll(dstPath); err != nil {
		return errors.Wrapf(err, "Could not delete old kyma source files in (%s)", dstPath)
	}
	repo, err := defaultCloner.Clone(url, dstPath, true)
	if err != nil {
		return errors.Wrapf(err, "Error downloading repository (%s)", url)
	}
	if rev != "" {
		return checkout(repo, url, rev)
	}
	return nil
}

type repoCloner interface {
	Clone(url, path string, noCheckout bool) (*git.Repository, error)
}

type remoteRepoCloner struct {
}

func (rc *remoteRepoCloner) Clone(url, path string, autoCheckout bool) (*git.Repository, error) {
	return git.PlainCloneContext(context.Background(), path, false, &git.CloneOptions{
		Depth:      0,
		URL:        url,
		NoCheckout: !autoCheckout,
	})
}

// revision can be 'main', a release version (e.g. 1.4.1), a commit hash (e.g. 34edf09a) or a PR (e.g. PR-9486).
func resolveRevision(repo *git.Repository, url, rev string) (*plumbing.Hash, error) {
	if strings.HasPrefix(rev, prPrefix) {
		fetchPR(repo, strings.TrimPrefix(rev, prPrefix)) // to ensure that the rev hash can be checked out
		err := error(nil)
		rev, err = resolvePRrevision(url, rev)
		if err != nil {
			return nil, err
		}
	}
	return repo.ResolveRevision(plumbing.Revision(rev))
}

func fetchPR(repo *git.Repository, prNmbr string) error {
	refs := []config.RefSpec{config.RefSpec(fmt.Sprintf("+refs/pull/%s/head:refs/remotes/origin/pr/%s", prNmbr, prNmbr))}
	return repo.Fetch(&git.FetchOptions{RefSpecs: refs})
}

func checkout(repo *git.Repository, url, rev string) error {
	w, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "Error getting the worktree")
	}
	hash, err := resolveRevision(repo, url, rev)
	if err != nil {
		return err
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash: *hash,
	})
	if err != nil {
		return errors.Wrap(err, "Error checking out revision")
	}
	return nil
}
