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
	Name  []string `(@Letter)+`                 // Directory name is now a sequence of letters
	Count *string  `(":" (@Number | @Letter))?` // Optional colon followed by a number or letter
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
	{"Letter", `[a-zA-Z]`}, // Single letters (used for both names and ranges)
	{"Colon", `:`},         // Colon by itself
	{"Number", `[0-9]+`},   // Numbers
	{"Comma", `,`},         // Comma separator
	{"GreaterThan", `>`},   // Greater-than symbol
	{"OpenBracket", `\[`},  // Opening bracket
	{"CloseBracket", `\]`}, // Closing bracket
	{"Whitespace", `\s+`},  // Whitespace
})

func prettyPrintLayout(layout *Layout, indent string) {
	if layout == nil {
		fmt.Println(indent + "nil")
		return
	}

	fmt.Println(indent + "Layout:")
	for i, part := range layout.Parts {
		fmt.Printf("%s  Part[%d]:\n", indent, i)
		prettyPrintPart(part, indent+"    ")
	}
}

func prettyPrintPart(part *Part, indent string) {
	if part == nil {
		fmt.Println(indent + "nil")
		return
	}

	if part.List != nil {
		fmt.Println(indent + "List:")
		for i, subLayout := range part.List.Layouts {
			fmt.Printf("%s  SubLayout[%d]:\n", indent, i)
			prettyPrintLayout(&subLayout, indent+"    ")
		}
	}
	if part.Level != nil {
		fmt.Printf("%sLevel: Name=%s", indent, part.Level.Name)
		if part.Level.Count != nil {
			fmt.Printf(", Count=%s", *part.Level.Count)
		}
		fmt.Println()
	}
}

// Build the parser with the updated grammar
var LayoutParser = participle.MustBuild[Layout](
	participle.Lexer(layoutLexer),
	participle.Elide("Whitespace"),
)

func main() {
	var input string
	var basePath string
	var debugTokens bool
	var debugParsing bool

	rootCmd := &cobra.Command{
		Use:   "dirlayout",
		Short: "Generate directory layouts using a simple syntax",
		Run: func(cmd *cobra.Command, args []string) {
			if input == "" {
				log.Fatal("Error: No layout string provided. Use the --layout flag to specify a layout.")
			}

			// Debugging: Lexer
			if debugTokens {
				debugLexer(input)
			}

			// Debugging: Parser
			if debugParsing {
				debugParser(input)
				return
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
	rootCmd.Flags().BoolVarP(&debugTokens, "debug-tokens", "t", false, "Enable tokenization debugging")
	rootCmd.Flags().BoolVarP(&debugParsing, "debug-parsing", "p", false, "Enable parser debugging")

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

func debugLexer(input string) {
	lex, _ := layoutLexer.LexString("", input)
	fmt.Println("Lexer Debugging Output:")
	for {
		token, err := lex.Next()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatalf("Lexer error: %v", err)
		}
		fmt.Printf("Token: %-10s Value: %q\n", token.Type, token.Value)
	}
	fmt.Println()
}

func debugParser(input string) {
	layout, err := LayoutParser.ParseString("", input)
	if err != nil {
		log.Fatalf("Parser error: %v", err)
	}

	fmt.Println("Parser Debugging Output:")
	prettyPrintLayout(layout, "")
}

func expandLevel(l *Level) []*DirectoryTree {
	// Concatenate letters to form the full name
	name := strings.Join(l.Name, "")

	if l.Count != nil {
		count := *l.Count
		var nodes []*DirectoryTree

		// Handle letter ranges (e.g., "a-f")
		if len(count) == 1 && count[0] >= 'a' && count[0] <= 'z' {
			for char := 'a'; char <= rune(count[0]); char++ {
				nodes = append(nodes, &DirectoryTree{
					Name:     fmt.Sprintf("%s %c", name, char),
					Children: nil,
				})
			}
			return nodes
		}

		// handle letter ranges (e.g., "A-F")
		if len(count) == 1 && count[0] >= 'A' && count[0] <= 'Z' {
			for char := 'A'; char <= rune(count[0]); char++ {
				nodes = append(nodes, &DirectoryTree{
					Name:     fmt.Sprintf("%s %c", name, char),
					Children: nil,
				})
			}
		}

		// Handle numeric ranges
		if c, err := strconv.Atoi(count); err == nil {
			for i := 1; i <= c; i++ {
				nodes = append(nodes, &DirectoryTree{
					Name:     fmt.Sprintf("%s %d", name, i),
					Children: nil,
				})
			}
			return nodes
		}

		log.Fatalf("Invalid count in level %q: %v", name, count)
	}

	// If no count, just a single directory
	return []*DirectoryTree{{
		Name:     name,
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
