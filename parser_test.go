package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_SingleLevelNameOnly(t *testing.T) {
	input := "hello"
	layout, err := LayoutParser.ParseString("", input)

	require.NoError(t, err)
	require.NotNil(t, layout)
	require.Len(t, layout.Parts, 1)

	firstPart := layout.Parts[0]
	require.NotNil(t, firstPart.Level)
	require.Nil(t, firstPart.List)

	require.Equal(t, "hello", firstPart.Level.Name)
	require.Nil(t, firstPart.Level.Count)
}

func TestParser_SingleLevelWithCount(t *testing.T) {
	input := "hello:3"
	layout, err := LayoutParser.ParseString("", input)

	require.NoError(t, err)
	require.NotNil(t, layout)
	require.Len(t, layout.Parts, 1)

	firstPart := layout.Parts[0]
	require.NotNil(t, firstPart.Level)
	require.Nil(t, firstPart.List)

	require.Equal(t, "hello", firstPart.Level.Name)
	require.NotNil(t, firstPart.Level.Count)
	require.Equal(t, "3", *firstPart.Level.Count)
}

func TestParser_MultipleLevels(t *testing.T) {
	input := "hello:2 > world > goodbye"
	layout, err := LayoutParser.ParseString("", input)

	require.NoError(t, err)
	require.NotNil(t, layout)
	require.Len(t, layout.Parts, 3)

	// Check first Part
	require.NotNil(t, layout.Parts[0].Level)
	require.Equal(t, "hello", layout.Parts[0].Level.Name)
	require.NotNil(t, layout.Parts[0].Level.Count)
	require.Equal(t, "2", *layout.Parts[0].Level.Count)

	// Check second Part
	require.NotNil(t, layout.Parts[1].Level)
	require.Equal(t, "world", layout.Parts[1].Level.Name)
	require.Nil(t, layout.Parts[1].Level.Count)

	// Check third Part
	require.NotNil(t, layout.Parts[2].Level)
	require.Equal(t, "goodbye", layout.Parts[2].Level.Name)
	require.Nil(t, layout.Parts[2].Level.Count)
}

func TestParser_BracketedListSimple(t *testing.T) {
	input := "[hello, goodbye]"
	layout, err := LayoutParser.ParseString("", input)

	require.NoError(t, err)
	require.NotNil(t, layout)
	require.Len(t, layout.Parts, 1)

	// The first part should be a list
	firstPart := layout.Parts[0]
	require.Nil(t, firstPart.Level)
	require.NotNil(t, firstPart.List)

	// Inside the list, we have multiple Layout entries
	require.Len(t, firstPart.List.Layouts, 2)

	// First sublayout should contain "hello"
	subLayout1 := firstPart.List.Layouts[0]
	require.Len(t, subLayout1.Parts, 1)
	require.NotNil(t, subLayout1.Parts[0].Level)
	require.Equal(t, "hello", subLayout1.Parts[0].Level.Name)

	// Second sublayout should contain "goodbye"
	subLayout2 := firstPart.List.Layouts[1]
	require.Len(t, subLayout2.Parts, 1)
	require.NotNil(t, subLayout2.Parts[0].Level)
	require.Equal(t, "goodbye", subLayout2.Parts[0].Level.Name)
}

func TestParser_BracketedListWithNestedLevels(t *testing.T) {
	input := "[hello:2 > world, goodbye] > universe"
	layout, err := LayoutParser.ParseString("", input)

	require.NoError(t, err)
	require.NotNil(t, layout)
	require.Len(t, layout.Parts, 2)

	// First part is a list
	listPart := layout.Parts[0]
	require.Nil(t, listPart.Level)
	require.NotNil(t, listPart.List)
	require.Len(t, listPart.List.Layouts, 2)

	// Check the first sublayout: "hello:2 > world"
	subLayout1 := listPart.List.Layouts[0]
	require.Len(t, subLayout1.Parts, 2)
	require.Equal(t, "hello", subLayout1.Parts[0].Level.Name)
	require.Equal(t, "2", *subLayout1.Parts[0].Level.Count)
	require.Equal(t, "world", subLayout1.Parts[1].Level.Name)

	// Check the second sublayout: "goodbye"
	subLayout2 := listPart.List.Layouts[1]
	require.Len(t, subLayout2.Parts, 1)
	require.Equal(t, "goodbye", subLayout2.Parts[0].Level.Name)

	// Second part is "universe"
	levelPart := layout.Parts[1]
	require.NotNil(t, levelPart.Level)
	require.Equal(t, "universe", levelPart.Level.Name)
	require.Nil(t, levelPart.Level.Count)
}

func TestParser_EmptyInput(t *testing.T) {
	input := ""
	layout, err := LayoutParser.ParseString("", input)

	require.Error(t, err)
	require.NotNil(t, layout)
	require.Empty(t, layout.Parts, "Expected no parts in an empty layout")
}

func TestParser_InvalidInput(t *testing.T) {
	input := "![invalid]" // Something outside the defined lexer rules
	_, err := LayoutParser.ParseString("", input)

	require.Error(t, err, "Parsing invalid input should produce an error")
}
