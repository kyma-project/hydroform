package downloader

import (
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/git"
)

const kymaURL = "https://github.com/kyma-project/kyma"

// CloneKymaRepo clones Kyma repo to the given dstPath
// version can be 'master', a release version (e.g. 1.4.1), a commit hash (e.g. 34edf09a) or a PR (e.g. PR-9486)
func CloneKymaRepo(dstPath, version string) error {
	rev, err := git.ResolveRevision(kymaURL, version)
	if err != nil {
		return err
	}

	if err := git.CloneRevision(kymaURL, dstPath, rev); err != nil {
		return err
	}

	return nil
}
