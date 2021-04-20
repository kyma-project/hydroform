package git

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const prPrefix = "PR-"

type client struct {
	lister lister
	cloner cloner
}

type lister interface {
	List(repoURL string) ([]*plumbing.Reference, error)
}

type remoteLister struct {
}

func (rl *remoteLister) List(repoURL string) ([]*plumbing.Reference, error) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	return remote.List(&git.ListOptions{})
}

type cloner interface {
	Clone(repoURL, dstPath string, noCheckout bool) (*git.Repository, error)
}

type remoteCloner struct {
}

func (rc *remoteCloner) Clone(repoURL, dstPath string, noCheckout bool) (*git.Repository, error) {
	return git.PlainCloneContext(context.Background(), dstPath, false, &git.CloneOptions{
		Depth:      0,
		URL:        repoURL,
		NoCheckout: noCheckout,
	})
}

var defaultClient = client{
	lister: &remoteLister{},
	cloner: &remoteCloner{},
}

// CloneRepo clones the repository in the given URL to the given dstPath and checks out the given revision.
// revision can be 'main', a release version (e.g. 1.4.1), a commit hash (e.g. 34edf09a) or a PR (e.g. PR-9486).
func CloneRepo(url, dstPath, rev string) error {
	return defaultClient.CloneRepo(url, dstPath, rev)
}

func (c *client) CloneRepo(url, dstPath, rev string) error {
	rev, err := c.ResolveRevision(url, rev)
	if err != nil {
		return err
	}

	if err := c.CloneRevision(url, dstPath, rev); err != nil {
		return err
	}

	return nil
}

// CloneRevision clones the repository in the given URL to the given dstPath and checks out the given revision.
// The clone downloads the bare minimum to only get the given revision.
// If the revision is empty, HEAD will be used.
func (c *client) CloneRevision(repoURL, dstPath, rev string) error {
	// only checkout HEAD if the revision is empty
	noCheckout := rev != ""
	r, err := c.cloner.Clone(repoURL, dstPath, noCheckout)
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

// ResolveRevision tries to convert a pseudo-revision reference (e.g. semVer, tag, PR, main, etc...) into a revision that can be checked out.
func (c *client) ResolveRevision(repo, rev string) (string, error) {
	switch {
	//Install the specific commit hash (e.g. 34edf09a)
	case isHex(rev):
		// no need for conversion
		return rev, nil

	//Install the specific version from release (ex: 1.15.1)
	case isSemVer(rev):
		// get tag commit ID
		return c.Tag(repo, rev)

	//Install the specific pull request (e.g. PR-9486)
	case strings.HasPrefix(rev, "PR-"):
		// get PR HEAD commit ID
		return c.PRHead(repo, rev)
	//Install the specific branch (e.g. main) or return error message
	default:
		if ref, err := c.BranchHead(repo, rev); err == nil {
			return ref, nil
		} else {
			return "", errors.Wrap(err, fmt.Sprintf("Could not find a branch with name '%s'\nfailed to parse the rev parameter. It can take one of the following: branch name (e.g. main), commit hash (e.g. 34edf09a), release version (e.g. 1.4.1), PR (e.g. PR-9486)", rev))
		}
	}
}

// BranchHead finds the HEAD commit hash of the given branch in the given repository.
func (c *client) BranchHead(repoURL, branch string) (string, error) {
	refs, err := c.lister.List(repoURL)
	if err != nil {
		return "", errors.Wrap(err, "could not list commits")
	}

	for _, ref := range refs {
		if ref.Name().IsBranch() && ref.Name().Short() == branch {
			return ref.Hash().String(), nil
		}
	}
	return "", errors.Errorf("could not find HEAD of branch %s in %s", branch, repoURL)
}

// Tag finds the commit hash of the given tag in the given repository.
func (c *client) Tag(repoURL, tag string) (string, error) {
	refs, err := c.lister.List(repoURL)
	if err != nil {
		return "", errors.Wrap(err, "could not list commits")
	}

	for _, ref := range refs {
		if ref.Name().IsTag() && ref.Name().Short() == tag {
			return ref.Hash().String(), nil
		}
	}
	return "", errors.Errorf("could not find tag %s in %s", tag, repoURL)
}

// PR finds the commit hash of the HEAD of the given PR in the given repository.
func (c *client) PRHead(repoURL, pr string) (string, error) {
	refs, err := c.lister.List(repoURL)
	if err != nil {
		return "", errors.Wrap(err, "could not list commits")
	}

	if strings.HasPrefix(pr, prPrefix) {
		pr = strings.TrimLeft(pr, prPrefix)
	}

	for _, ref := range refs {
		if strings.HasPrefix(ref.Name().String(), "refs/pull") && strings.HasSuffix(ref.Name().String(), "head") && strings.Contains(ref.Name().String(), pr) {
			return ref.Hash().String(), nil
		}
	}
	return "", errors.Errorf("could not find HEAD of pull request %s in %s", pr, repoURL)
}

func isSemVer(s string) bool {
	_, err := semver.Parse(s)
	return err == nil
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil && len(s) > 7
}
