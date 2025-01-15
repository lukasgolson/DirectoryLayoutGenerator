package main

import (
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Layout struct {
	Levels []*Level `@@ ( ">" @@ )*`
}

type Level struct {
	Name  string     `(@Ident | @Number)?`
	Count *string    `( ":" @Number )?`
	List  *ValueList `( @@ )?`
}

type Value struct {
	Name  string  `(@Ident | @Number)` // e.g. "lol" or "foo" or "123"
	Count *string `( ":" @Number )?`   // e.g. ":2" if present
}

type ValueList struct {
	Values []Value `"[" @@ ( "," @@ )* "]"`
}

type DirectoryTree struct {
	Name     string
	Children []*DirectoryTree
}

var layoutLexer = lexer.MustSimple([]lexer.SimpleRule{
	{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},
	{"Number", `[0-9]+`},
	{"Colon", `:`},
	{"Comma", `,`},
	{"GreaterThan", `>`},
	{"OpenBracket", `\[`},
	{"CloseBracket", `\]`},
	{"Whitespace", `\s+`},
})

var layoutParser = participle.MustBuild[Layout](
	participle.Lexer(layoutLexer),
	participle.Elide("Whitespace"),
)

func main() {
	var input string
	var basePath string

	rootCmd := &cobra.Command{
		Use:   "dirlayout",
		Short: "Generate directory layouts using a simple syntax",
		Run: func(cmd *cobra.Command, args []string) {
			if input == "" {
				log.Fatal("Error: No layout string provided. Use the --layout flag to specify a layout.")
			}

			parsedLayout, err := layoutParser.ParseString("", input)
			if err != nil {
				log.Fatalf("Error parsing layout: %v", err)
			}

			tree := buildDirectoryTree(parsedLayout.Levels, 0)
			if err := createDirectoryTree(basePath, tree); err != nil {
				log.Fatalf("Error creating directories: %v", err)
			}

			fmt.Println("Directory layout created successfully!")
		},
	}

	rootCmd.Flags().StringVarP(&input, "layout", "l", "", "Layout string describing the directory structure (e.g., 'site:5 > tree:10')")
	rootCmd.Flags().StringVarP(&basePath, "output", "o", ".", "Base path where the directories will be created")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func buildDirectoryTree(levels []*Level, levelIndex int) *DirectoryTree {
	if levelIndex >= len(levels) {
		return nil
	}

	level := levels[levelIndex]
	var names []string

	// Collect folder names for THIS level
	switch {
	case level.List != nil:

		for _, v := range level.List.Values {
			if v.Count != nil {
				count, err := strconv.Atoi(*v.Count)
				if err != nil {
					log.Fatalf("Invalid count in array: %v", err)
				}
				for i := 1; i <= count; i++ {
					names = append(names, fmt.Sprintf("%s %d", v.Name, i))
				}
			} else {
				names = append(names, v.Name)
			}
		}

	case level.Count != nil:
		// e.g. "three:2" => ["three 1", "three 2"]
		count, err := strconv.Atoi(*level.Count)
		if err != nil {
			log.Fatalf("Invalid count: %v", err)
		}
		for i := 1; i <= count; i++ {
			names = append(names, fmt.Sprintf("%s %d", level.Name, i))
		}

	case level.Name != "":
		names = []string{level.Name}
	}

	// A "container" node to hold children
	container := &DirectoryTree{
		Name:     "",
		Children: []*DirectoryTree{},
	}

	// For each name, create a child node and attach subTree's children
	for _, n := range names {
		child := &DirectoryTree{
			Name:     n,
			Children: []*DirectoryTree{},
		}

		// Recursively build next level
		subTree := buildDirectoryTree(levels, levelIndex+1)
		if subTree != nil {
			// DO NOT overwrite subTree.Name!  Just attach its children.
			child.Children = subTree.Children
		}

		// Add child to the container
		container.Children = append(container.Children, child)
	}

	return container
}

func createDirectoryTree(basePath string, tree *DirectoryTree) error {
	if tree == nil {
		return nil
	}

	currentPath := filepath.Join(basePath, strings.TrimSpace(tree.Name))
	if tree.Name != "" {
		if err := os.MkdirAll(currentPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", currentPath, err)
		}
	}

	for _, child := range tree.Children {
		if err := createDirectoryTree(currentPath, child); err != nil {
			return err
		}
	}

	return nil
}
