package git

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type fakeRefLister struct {
	refs []*plumbing.Reference
}

func (fl *fakeRefLister) List(repoURL string) ([]*plumbing.Reference, error) {
	return fl.refs, nil
}

// TestResolveRevision tests implicitly also the commit ID resolution functions for: Branch, PR and Tag
func TestResolveRevision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		summary       string
		givenRefs     []*plumbing.Reference
		givenRevision string
		expectErr     bool
	}{
		{
			summary: "main branch head",
			givenRefs: []*plumbing.Reference{
				plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), plumbing.ZeroHash),
				plumbing.NewHashReference(plumbing.NewTagReferenceName("1.0"), plumbing.ZeroHash),
			},
			givenRevision: "main",
		},
		{
			summary: "semver tag",
			givenRefs: []*plumbing.Reference{
				plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), plumbing.ZeroHash),
				plumbing.NewHashReference(plumbing.NewTagReferenceName("1.15.0"), plumbing.ZeroHash),
			},
			givenRevision: "1.15.0",
		},
		{
			summary: "pull request uppercase",
			givenRefs: []*plumbing.Reference{
				plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), plumbing.ZeroHash),
				plumbing.NewHashReference(plumbing.NewTagReferenceName("1.0"), plumbing.ZeroHash),
				plumbing.NewHashReference(plumbing.ReferenceName("refs/pull/9999/head"), plumbing.ZeroHash),
			},
			givenRevision: "PR-9999",
		},
		{
			summary: "bad ref",
			givenRefs: []*plumbing.Reference{
				plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), plumbing.ZeroHash),
				plumbing.NewHashReference(plumbing.NewTagReferenceName("1.0"), plumbing.ZeroHash),
				plumbing.NewHashReference(plumbing.ReferenceName("refs/pull/9999/head"), plumbing.ZeroHash),
			},
			givenRevision: "not-a-git-ref",
			expectErr:     true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.summary, func(t *testing.T) {
			resolver := revisionResolver{
				lister: &fakeRefLister{
					refs: tc.givenRefs,
				},
			}
			r, err := resolver.resolveRevision(repo, tc.givenRevision)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, isHex(r))
			}
		})
	}
}
