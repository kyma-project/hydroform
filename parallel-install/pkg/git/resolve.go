package git

import (
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

type revisionResolver struct {
	lister refLister
}

type refLister interface {
	List(repoURL string) ([]*plumbing.Reference, error)
}

type remoteRefLister struct {
}

func (rl *remoteRefLister) List(repoURL string) ([]*plumbing.Reference, error) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	return remote.List(&git.ListOptions{})
}

var defaultResolver = revisionResolver{
	lister: &remoteRefLister{},
}

// resolveRevision tries to convert a pseudo-revision reference (e.g. semVer, tag, PR, main, etc...) into a revision that can be checked out.
func (c *revisionResolver) resolveRevision(repo, rev string) (string, error) {
	switch {
	// Install the specific commit hash (e.g. 34edf09a)
	case isHex(rev):
		// no need for conversion
		return rev, nil

	// Install the specific version from release (ex: 1.15.1)
	case isSemVer(rev):
		// get tag commit ID
		return c.tag(repo, rev)

	// Install the specific pull request (e.g. PR-9486)
	case strings.HasPrefix(rev, "PR-"):
		// get PR HEAD commit ID
		return c.prHead(repo, rev)
	// Install the specific branch (e.g. main) or return error message
	default:
		if ref, err := c.branchHead(repo, rev); err == nil {
			return ref, nil
		} else {
			return "", errors.Wrap(err, fmt.Sprintf("Could not find a branch with name '%s'\nfailed to parse the rev parameter. It can take one of the following: branch name (e.g. main), commit hash (e.g. 34edf09a), release version (e.g. 1.4.1), PR (e.g. PR-9486)", rev))
		}
	}
}

// branchHead finds the HEAD commit hash of the given branch in the given repository.
func (c *revisionResolver) branchHead(repoURL, branch string) (string, error) {
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

// tag finds the commit hash of the given tag in the given repository.
func (c *revisionResolver) tag(repoURL, tag string) (string, error) {
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
func (c *revisionResolver) prHead(repoURL, pr string) (string, error) {
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
