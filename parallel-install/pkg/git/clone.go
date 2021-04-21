package git

import (
	"context"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type repoCloner interface {
	Clone(url, path string, noCheckout bool) (*git.Repository, error)
}

type remoteRepoCloner struct {
}

func (rc *remoteRepoCloner) Clone(url, path string, noCheckout bool) (*git.Repository, error) {
	return git.PlainCloneContext(context.Background(), path, false, &git.CloneOptions{
		Depth:      0,
		URL:        url,
		NoCheckout: noCheckout,
	})
}

var defaultCloner repoCloner = &remoteRepoCloner{}

// CloneRepo clones the repository in the given URL to the given dstPath and checks out the given revision.
// revision can be 'main', a release version (e.g. 1.4.1), a commit hash (e.g. 34edf09a) or a PR (e.g. PR-9486).
func CloneRepo(url, dstPath, rev string) error {
	rev, err := defaultResolver.resolveRevision(url, rev)
	if err != nil {
		return err
	}

	if err := CloneRevision(url, dstPath, rev); err != nil {
		return err
	}

	return nil
}

// CloneRevision clones the repository in the given URL to the given dstPath and checks out the given revision.
// The clone downloads the bare minimum to only get the given revision.
// If the revision is empty, HEAD will be used.
func CloneRevision(repoURL, dstPath, rev string) error {
	// only checkout HEAD if the revision is empty
	noCheckoutWhenClone := rev != ""
	r, err := defaultCloner.Clone(repoURL, dstPath, noCheckoutWhenClone)
	if err != nil {
		return errors.Wrapf(err, "Error downloading repository (%s)", repoURL)
	}

	if noCheckoutWhenClone {
		w, err := r.Worktree()
		if err != nil {
			return errors.Wrap(err, "Error getting the worktree")
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(rev),
		})
		if err != nil {
			return errors.Wrap(err, "Error checking out revision")
		}
	}
	return nil
}
