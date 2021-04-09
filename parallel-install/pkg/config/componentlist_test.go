package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ComponentList_New(t *testing.T) {
	t.Run("From YAML", func(t *testing.T) {
		newCompList(t, "../test/data/componentlist.yaml")
	})
	t.Run("From JSON", func(t *testing.T) {
		newCompList(t, "../test/data/componentlist.json")
	})
}

func Test_ComponentList_Remove(t *testing.T) {
	t.Run("Remove Prerequisite", func(t *testing.T) {
		compList := newCompList(t, "../test/data/componentlist.yaml")
		compList.Remove("prereqcomp1")
		require.Equal(t, 1, len(compList.Prerequisites), "Different amount of prerequisite components")
		require.Equal(t, 3, len(compList.Components), "Different amount of components")
		require.Equal(t, "prereqcomp2", compList.Prerequisites[0].Name)
	})
	t.Run("Remove Component", func(t *testing.T) {
		compList := newCompList(t, "../test/data/componentlist.yaml")
		compList.Remove("comp2")
		require.Equal(t, 2, len(compList.Prerequisites), "Different amount of prerequisite components")
		require.Equal(t, 2, len(compList.Components), "Different amount of components")
		require.Equal(t, "comp1", compList.Components[0].Name)
		require.Equal(t, "comp3", compList.Components[1].Name)
	})
}

func Test_ComponentList_Add(t *testing.T) {
	t.Run("Add Component in default namespace", func(t *testing.T) {
		compList := newCompList(t, "../test/data/componentlist.yaml")
		compList.Add("comp4", "")
		require.Equal(t, "comp4", compList.Components[3].Name)
		require.Equal(t, defaultNamespace, compList.Components[3].Namespace)
	})
	t.Run("Add Component in custom namespace", func(t *testing.T) {
		compList := newCompList(t, "../test/data/componentlist.yaml")
		namespace := "test-namespace"
		compList.Add("comp4", namespace)
		require.Equal(t, "comp4", compList.Components[3].Name)
		require.Equal(t, namespace, compList.Components[3].Namespace)
	})
}

func verifyComponentList(t *testing.T, compList *ComponentList) {
	prereqs := compList.Prerequisites
	comps := compList.Components
	// verify amount of components

	require.Equal(t, 2, len(prereqs), "Different amount of prerequisite components")
	require.Equal(t, 3, len(comps), "Different amount of components")

	// verify names + namespaces of prerequisites
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

func newCompList(t *testing.T, compFile string) *ComponentList {
	compList, err := NewComponentList(compFile)
	require.NoError(t, err)
	verifyComponentList(t, compList)
	return compList
}
