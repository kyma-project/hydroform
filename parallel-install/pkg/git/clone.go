package git

import (
	"context"

	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

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
	r, err := git.PlainCloneContext(context.Background(), dstPath, false, &git.CloneOptions{
		Depth:      0,
		URL:        repoURL,
		NoCheckout: rev != "", // only checkout HEAD if the revision is empty
	})
	if err != nil {
		return errors.Wrapf(err, "Error downloading repository (%s)", repoURL)
	}

	if rev != "" {
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
