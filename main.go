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
)

type Layout struct {
	Levels []*Level `@@ ( ">" @@ )*`
}

type Level struct {
	Name   string   `@Ident`
	Count  *string  `( ":" @Number )?`
	Values []string `( ":" @Ident ( "," @Ident )* )?`
}

var layoutLexer = lexer.MustSimple([]lexer.SimpleRule{
	{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},
	{"Number", `[0-9]+`},
	{"Colon", `:`},
	{"Comma", `,`},
	{"GreaterThan", `>`},
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
		Long:  `dirlayout is a CLI tool for generating complex directory structures using a simple custom domain-specific language called the Directory Structure Language (creative, I know).".`,
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

	if level.Count != nil {
		count, err := strconv.Atoi(*level.Count) // Convert *string to int
		if err != nil {
			return fmt.Errorf("invalid count: %v", err)
		}
		for i := 1; i <= count; i++ {
			names = append(names, fmt.Sprintf("%s %d", level.Name, i))
		}
	} else if len(level.Values) > 0 {
		names = level.Values
	} else {
		names = []string{level.Name}
	}

	for _, name := range names {
		fullPath := filepath.Join(basePath, name)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", fullPath, err)
		}
		if err := createDirectories(fullPath, levels, levelIndex+1); err != nil {
			return err
		}
	}
	return nil
}
