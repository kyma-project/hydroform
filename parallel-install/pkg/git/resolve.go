package git

import (
	"encoding/hex"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

const prPrefix = "PR-"

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

var defaultLister refLister = &remoteRefLister{}

// branchHead finds the HEAD commit hash of the given branch in the given repository.
func BranchHead(repoURL, branch string) (string, error) {
	refs, err := defaultLister.List(repoURL)
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
func Tag(repoURL, tag string) (string, error) {
	refs, err := defaultLister.List(repoURL)
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

// resolvePRrevision tries to convert a PR into a revision that can be checked out.
func resolvePRrevision(repoURL, pr string) (string, error) {
	refs, err := defaultLister.List(repoURL)
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
