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
	Parts []*Part `@@ ( ">" @@ )*`
}

type Part struct {
	List *ValueList `  @@ 
                     | `
	Level *Level `  @@ `
}

type Level struct {
	Name  string  `(@Ident | @Number)` // <-- no question mark
	Count *string `( ":" @Number )?`
}

type ValueList struct {
	Layouts []Layout `"[" @@ ( "," @@ )* "]"`
}

type DirectoryTree struct {
	Name     string
	Children []*DirectoryTree
}

// Our lexer:
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

// Build the parser with the updated grammar
var LayoutParser = participle.MustBuild[Layout](
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

			// Build an in-memory directory tree
			tree, err := ParseAndBuildDirectoryTree(input)

			if err != nil {
				log.Fatalf("Error parsing layout: %v", err)
			}

			// Create directories on disk
			if err := createDirectoryTree(basePath, tree); err != nil {
				log.Fatalf("Error creating directories: %v", err)
			}

			fmt.Println("Directory layout created successfully!")
		},
	}

	rootCmd.Flags().StringVarP(&input, "layout", "l", "",
		`Layout string describing the directory structure (e.g., "[lol:2 > lukas, lmfao] > test")`,
	)
	rootCmd.Flags().StringVarP(&basePath, "output", "o", ".",
		"Base path where the directories will be created",
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ParseAndBuildDirectoryTree(input string) (*DirectoryTree, error) {
	// 1) Parse the input into a Layout structure
	layout, err := LayoutParser.ParseString("", input)
	if err != nil {
		return nil, err
	}

	// 2) Build the directory tree
	tree := buildLayoutTree(*layout)

	return tree, nil
}

// buildLayoutTree creates a DirectoryTree from a Layout *recursively*,
// interpreting each ">" as a deeper nesting.
func buildLayoutTree(layout Layout) *DirectoryTree {
	// If there are no parts at all, return an empty container
	if len(layout.Parts) == 0 {
		return &DirectoryTree{Name: "", Children: nil}
	}

	// 1) Build directories for the *first* part
	firstPart := layout.Parts[0]
	topNodes := buildPart(firstPart)

	// 2) If there are more parts, recursively build the subtree for them
	if len(layout.Parts) > 1 {
		rest := Layout{Parts: layout.Parts[1:]}
		restTree := buildLayoutTree(rest)

		// Attach restTree's children as subdirectories of each node from topNodes
		for _, node := range topNodes {
			node.Children = append(node.Children, restTree.Children...)
		}
	}

	// 3) Return a container whose direct children are these topNodes
	return &DirectoryTree{Name: "", Children: topNodes}
}

// buildPart handles a single Part (which can be a Level or a bracketed ValueList).
func buildPart(part *Part) []*DirectoryTree {
	if part.Level != nil {
		// E.g. "lol:2"
		return expandLevel(part.Level)
	}
	if part.List != nil {
		// Bracketed lists, e.g. "[lol:2 > lukas, lmfao]"
		// Each comma-separated element inside is itself a Layout
		// We'll build each sub-Layout and then combine them as siblings.
		var result []*DirectoryTree
		for _, subLayout := range part.List.Layouts {
			subTree := buildLayoutTree(subLayout)
			// subTree is an empty container with children
			result = append(result, subTree.Children...)
		}
		return result
	}
	// Should never happen if grammar is correct
	return nil
}

// expandLevel returns 1 or more DirectoryTree nodes for "Name(:Count)?"
func expandLevel(l *Level) []*DirectoryTree {
	if l.Count != nil {
		c, err := strconv.Atoi(*l.Count)
		if err != nil {
			log.Fatalf("Invalid count in level %q: %v", l.Name, err)
		}
		var nodes []*DirectoryTree
		for i := 1; i <= c; i++ {
			nodes = append(nodes, &DirectoryTree{
				Name:     fmt.Sprintf("%s %d", l.Name, i),
				Children: nil,
			})
		}
		return nodes
	}

	// If no count, just a single directory
	return []*DirectoryTree{{
		Name:     l.Name,
		Children: nil,
	}}
}

// Recursively create the directories on disk
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
