package helm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_VersionComponentSorting(t *testing.T) {
	version := &KymaVersion{
		components: []*KymaComponentMetadata{
			&KymaComponentMetadata{ //No. 4
				Name:         "4",
				Priority:     12,
				Prerequisite: false,
			},
			&KymaComponentMetadata{ //No. 2
				Name:         "2",
				Priority:     0,
				Prerequisite: false,
			},
			&KymaComponentMetadata{ //No. 1
				Name:         "1",
				Priority:     99,
				Prerequisite: true,
			},
			&KymaComponentMetadata{ //No. 3
				Name:         "3",
				Priority:     2,
				Prerequisite: false,
			},
			&KymaComponentMetadata{ //No. 0
				Name:         "0",
				Priority:     18,
				Prerequisite: true,
			},
		},
	}
	sortedComps := version.Components()
	for idx, comp := range sortedComps {
		require.Equal(t, fmt.Sprintf("%d", idx), comp.Name)
	}
}

func Test_VersionSetComponentSorting(t *testing.T) {
	versionSet := &KymaVersionSet{
		Versions: []*KymaVersion{
			&KymaVersion{
				components: []*KymaComponentMetadata{
					&KymaComponentMetadata{ //No. 6
						Name:         "6",
						Priority:     12,
						Prerequisite: false,
					},
					&KymaComponentMetadata{ //No. 5
						Name:         "5",
						Priority:     2,
						Prerequisite: false,
					},
					&KymaComponentMetadata{ //No. 1
						Name:         "1",
						Priority:     18,
						Prerequisite: true,
					},
					&KymaComponentMetadata{ //No. 0
						Name:         "0",
						Priority:     7,
						Prerequisite: true,
					},
				},
			},
			&KymaVersion{
				components: []*KymaComponentMetadata{
					&KymaComponentMetadata{ //No. 4
						Name:         "4",
						Priority:     0,
						Prerequisite: false,
					},
					&KymaComponentMetadata{ //No. 2
						Name:         "2",
						Priority:     99,
						Prerequisite: true,
					},
					&KymaComponentMetadata{ //No. 3
						Name:         "3",
						Priority:     0,
						Prerequisite: false,
					},
				},
			},
		},
	}
	sortedComps := versionSet.Components()
	for idx, comp := range sortedComps {
		require.Equal(t, fmt.Sprintf("%d", idx), comp.Name)
	}
}
