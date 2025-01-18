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

// buildLayoutTree creates a DirectoryTree from a Layout *recursively*,
// interpreting each ">" as a deeper nesting.
func buildLayoutTree(layout Layout) *DirectoryTree {
	if len(layout.Parts) == 0 {
		return &DirectoryTree{Name: "", Children: nil}
	}

	// Build directories for the first part.
	firstPart := layout.Parts[0]
	topNodes := buildPart(firstPart)

	// If there are more parts, recursively build the subtree for them.
	if len(layout.Parts) > 1 {
		rest := Layout{Parts: layout.Parts[1:]}
		restTree := buildLayoutTree(rest)

		// Attach restTree's children to every leaf node in topNodes.
		for _, node := range topNodes {
			attachToLeaves(node, restTree.Children)
		}
	}

	return &DirectoryTree{Name: "", Children: topNodes}
}

// buildPart handles a single Part (which can be a Level or a bracketed ValueList).
func buildPart(part *Part) []*DirectoryTree {
	if part.Level != nil {
		// E.g., "lol:2"
		return expandLevel(part.Level)
	}
	if part.List != nil {
		// Bracketed lists, e.g., "[lukas > test2, test3]"
		// Each comma-separated element inside is itself a Layout.
		// We'll build each sub-Layout and then combine them as siblings.
		var result []*DirectoryTree
		for _, subLayout := range part.List.Layouts {
			// Each subLayout is treated as its own top-level tree
			subTree := buildLayoutTree(subLayout)
			result = append(result, subTree.Children...)
		}
		return result
	}
	// Should never happen if grammar is correct
	return nil
}

// attachToLeaves recursively finds all leaf nodes in the tree and attaches children to them.
func attachToLeaves(tree *DirectoryTree, children []*DirectoryTree) {
	if len(tree.Children) == 0 {
		tree.Children = append(tree.Children, cloneTreeList(children)...)
	} else {
		for _, child := range tree.Children {
			attachToLeaves(child, children)
		}
	}
}

// cloneTreeList creates a deep copy of the children list to avoid circular references.
func cloneTreeList(children []*DirectoryTree) []*DirectoryTree {
	var cloned []*DirectoryTree
	for _, child := range children {
		cloned = append(cloned, &DirectoryTree{
			Name:     child.Name,
			Children: cloneTreeList(child.Children), // Recursively clone children.
		})
	}
	return cloned
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
