// tree_test.go
package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildLayoutTree_Simple(t *testing.T) {
	// If you want to bypass the parser, you can manually create
	// a Layout object for this test.
	layout := Layout{
		Parts: []*Part{
			{
				Level: &Level{Name: "hello"},
			},
			{
				Level: &Level{Name: "world", Count: strPtr("2")},
			},
		},
	}

	// Call buildLayoutTree directly
	tree := buildLayoutTree(layout)

	// Verify top-level node is empty container
	require.Equal(t, "", tree.Name)
	require.Len(t, tree.Children, 1, "Expected one child: 'hello'")

	// The child "hello" should have the children from the second part "world:2"
	helloNode := tree.Children[0]
	require.Equal(t, "hello", helloNode.Name)
	require.Len(t, helloNode.Children, 2, "Expected two subdirectories named 'world 1' and 'world 2'")
	require.Equal(t, "world 1", helloNode.Children[0].Name)
	require.Equal(t, "world 2", helloNode.Children[1].Name)
}

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
