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
	List  *ValueList `( ">"? @@ )?`
}

type ValueList struct {
	Values []string `"[" (@Ident | @Number) ( "," (@Ident | @Number) )* "]"`
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
		Long: `dirlayout is a CLI tool for generating complex directory structures 
using a custom domain-specific language called the Directory Structure Language (creative, I know).

Basic Syntax:
1. Directories:
   - Use 'name[:count]', where:
     - 'name' is the base name for directories.
     - 'count' (optional) specifies the number of directories to create with that name.
       If 'count' is omitted, a single directory will be created.
     - Example: 'site:5' creates 5 directories named "site".

2. Nesting:
   - Represent nesting using the '>' symbol.
   - Example: 'site:5 > tree:10' creates 5 directories named "site",
     each containing 10 directories named "tree".

3. Static Lists:
   - Use square brackets '[]' to provide static lists of directory names.
   - Example: 'site:5 > tree:10 > [a, b, c]' creates:
     - 5 directories named "site", each containing:
     - 10 directories named "tree", each containing:
     - Directories named "a", "b", and "c".`,

		Run: func(cmd *cobra.Command, args []string) {
			if input == "" {
				log.Fatal("Error: No layout string provided. Use the --layout flag to specify a layout.")
			}

			layout, err := layoutParser.ParseString("", input)
			if err != nil {
				log.Fatalf("Error parsing layout: %v", err)
			}

			if err := createDirectories(basePath, layout.Levels, 0); err != nil {
				log.Fatalf("Error creating directories: %v", err)
			}

			fmt.Println("Directory layout created successfully!")
		},
	}

	// Add flags
	rootCmd.Flags().StringVarP(&input, "layout", "l", "", "Layout string describing the directory structure (e.g., 'site:5 > tree:10')")
	rootCmd.Flags().StringVarP(&basePath, "output", "o", ".", "Base path where the directories will be created")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createDirectories(basePath string, levels []*Level, levelIndex int) error {
	if levelIndex >= len(levels) {
		return nil
	}

	level := levels[levelIndex]
	var names []string

	// Handle directory naming based on count or static lists
	if level.List != nil {
		names = level.List.Values
	} else if level.Count != nil {
		// Parse count and generate numbered names
		count, err := strconv.Atoi(*level.Count)
		if err != nil {
			return fmt.Errorf("invalid count: %v", err)
		}
		for i := 1; i <= count; i++ {
			names = append(names, fmt.Sprintf("%s %d", level.Name, i))
		}
	} else if level.Name != "" {
		// Default to using the level name if no count or list is provided
		names = []string{level.Name}
	}

	// Always create the current directory level, even if it has a ValueList
	currentBasePath := basePath
	if level.Name != "" {
		currentBasePath = filepath.Join(basePath, strings.TrimSpace(level.Name))
		if err := os.MkdirAll(currentBasePath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", currentBasePath, err)
		}
	}

	// Create subdirectories and recurse for nested levels
	for _, name := range names {
		fullPath := filepath.Join(currentBasePath, strings.TrimSpace(name))
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", fullPath, err)
		}
		// Recurse into the next level
		if err := createDirectories(fullPath, levels, levelIndex+1); err != nil {
			return err
		}
	}

	return nil
}
