package components

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CLGetComponents(t *testing.T) {
	clList, err := NewComponentList("../test/data/componentslist.yaml")
	require.NoError(t, err)

	prereqs := clList.GetPrerequisites()
	comps := clList.GetComponents()
	// verify amount of components

	require.Equal(t, 2, len(prereqs), "Different amount of prerequisite components")
	require.Equal(t, 3, len(comps), "Different amount of components")

	// verify names + namespaces of prerequisistes
	require.Equal(t, "prereqcomp1", prereqs[0].Name, "Wrong component name")
	require.Equal(t, "prereqns1", prereqs[0].Namespace, "Wrong namespace")
	require.Equal(t, "prereqcomp2", prereqs[1].Name, "Wrong component name")
	require.Equal(t, "testns", prereqs[1].Namespace, "Wrong namespace")

	// verify names + namespaces of components
	require.Equal(t, "comp1", comps[0].Name, "Wrong component name")
	require.Equal(t, "testns", comps[0].Namespace, "Wrong namespace")
	require.Equal(t, "comp2", comps[1].Name, "Wrong component name")
	require.Equal(t, "compns2", comps[1].Namespace, "Wrong namespace")
	require.Equal(t, "comp3", comps[2].Name, "Wrong component name")
	require.Equal(t, "testns", comps[2].Namespace, "Wrong namespace")
}
