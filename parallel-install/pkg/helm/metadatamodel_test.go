package helm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_KymaVersionSet(t *testing.T) {
	t.Run("Latest version", func(t *testing.T) {
		//we abuse the name to verify the correct sequence
		versionSet := &KymaVersionSet{
			Versions: []*KymaVersion{
				{
					Version:      "old",
					CreationTime: 11111,
				},
				{
					Version:      "latest",
					CreationTime: 33333,
				},
				{
					Version:      "middle",
					CreationTime: 23456,
				},
			},
		}
		require.Equal(t, "latest", versionSet.Latest().Version)
	})

	t.Run("Sort components", func(t *testing.T) {
		//we abuse the name to verify the correct sequence
		versionSet := &KymaVersionSet{
			Versions: []*KymaVersion{
				{
					Components: []*KymaComponentMetadata{
						{ //No. 6
							Name:         "6",
							Priority:     12,
							Prerequisite: false,
						},
						{ //No. 5
							Name:         "5",
							Priority:     2,
							Prerequisite: false,
						},
						{ //No. 1
							Name:         "1",
							Priority:     18,
							Prerequisite: true,
						},
						{ //No. 0
							Name:         "0",
							Priority:     7,
							Prerequisite: true,
						},
					},
				},
				{
					Components: []*KymaComponentMetadata{
						{ //No. 4
							Name:         "4",
							Priority:     0,
							Prerequisite: false,
						},
						{ //No. 2
							Name:         "2",
							Priority:     99,
							Prerequisite: true,
						},
						{ //No. 3
							Name:         "3",
							Priority:     0,
							Prerequisite: false,
						},
					},
				},
			},
		}
		sortedComps := versionSet.InstalledComponents()
		for idx, comp := range sortedComps {
			require.Equal(t, fmt.Sprintf("%d", idx), comp.Name) //expected order position is reflected in name
		}
	})
}
func Test_KymaVersion(t *testing.T) {
	t.Run("Sort components", func(t *testing.T) {
		//we abuse the name to verify the correct sequence
		version := &KymaVersion{
			Components: []*KymaComponentMetadata{
				{ //No. 4
					Name:         "4",
					Priority:     12,
					Prerequisite: false,
				},
				{ //No. 2
					Name:         "2",
					Priority:     0,
					Prerequisite: false,
				},
				{ //No. 1
					Name:         "1",
					Priority:     99,
					Prerequisite: true,
				},
				{ //No. 3
					Name:         "3",
					Priority:     2,
					Prerequisite: false,
				},
				{ //No. 0
					Name:         "0",
					Priority:     18,
					Prerequisite: true,
				},
			},
		}
		sortedComps := version.InstalledComponents()
		for idx, comp := range sortedComps {
			require.Equal(t, fmt.Sprintf("%d", idx), comp.Name) //expected order position is reflected in name
		}
	})
}
