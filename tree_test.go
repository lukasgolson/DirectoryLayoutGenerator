// tree_test.go
package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildLayoutTree_OneLevel(t *testing.T) {
	input := "hello"

	tree, err := ParseAndBuildDirectoryTree(input)
	require.NoError(t, err)

	require.Len(t, tree.Children, 1)
	require.Equal(t, "hello", tree.Children[0].Name)
}

func TestBuildLayoutTree_OneLevelWithCount(t *testing.T) {
	input := "hello:3"

	tree, err := ParseAndBuildDirectoryTree(input)
	require.NoError(t, err)

	require.Len(t, tree.Children, 3)
	for i := 1; i <= 3; i++ {

		expectedString := fmt.Sprintf("hello %d", i)

		require.Equal(t, expectedString, tree.Children[i-1].Name)
	}
}

func TestBuildLayoutTree_NumericNames(t *testing.T) {
	input := "1:3"

	tree, err := ParseAndBuildDirectoryTree(input)

	require.NoError(t, err)

	require.Len(t, tree.Children, 3)

	for i := 1; i <= 3; i++ {
		expectedString := fmt.Sprintf("1 %d", i)
		require.Equal(t, expectedString, tree.Children[i-1].Name)
	}

}

func TestBuildLayoutTree_BracketsAndLevels(t *testing.T) {
	input := "[hello:2 > world, earth] > test"

	tree, err := ParseAndBuildDirectoryTree(input)
	require.NoError(t, err)

	// Top-level container node should have 3 children, hello 1, hello 2, and earth.
	require.Len(t, tree.Children, 3)

	// under hello 1 and hello 2, there should be a world directory

	for _, child := range tree.Children {
		require.Len(t, child.Children, 1)
		require.Equal(t, "world", child.Children[0].Name)
	}

	// And each child has a "test" subdirectory.
	for _, child := range tree.Children {
		require.Len(t, child.Children, 1)
		require.Equal(t, "test", child.Children[0].Name)
	}
}

func strPtr(s string) *string {
	return &s
}
