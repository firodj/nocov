package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/cover"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func usage() {
	//nocoverage
	fmt.Fprintf(os.Stderr, "nocov - a golang tool to remove blocks of code from the coverage report\n\n")
	flag.PrintDefaults()
}

func main() {

	//nocoverage
	gopath := os.Getenv("GOPATH")

	var coverFilename string
	var commentMarker string
	var coverCount int
	flag.StringVar(&coverFilename, "coverprofile", "c.out", "The filename of the cover profile as generated by : go test -coverprofile=<coverprofile>")
	flag.StringVar(&commentMarker, "commentMarker", "//nocoverage", "The comment to search for in the code that is used to define a block that cannot be covered by the tests")
	flag.IntVar(&coverCount, "coverCount", -1, "Blocks marked with the <commentMarker> will be assigned to have been covered <coverCount> times. If this number is negative, the block is removed from the coverage report")

	flag.Usage = usage
	flag.Parse()

	profiles, err := cover.ParseProfiles(coverFilename)
	if err != nil {
		//nocoverage
		log.Fatalf("Failed to open cover profile file: %s: %s", coverFilename, err)
		os.Exit(1)
	}

	for _, profile := range profiles {

		//nocoverage

		fileName := filepath.Join(gopath, "src", profile.FileName)

		fset := token.NewFileSet()
		parsedFile, err := parser.ParseFile(fset, fileName, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("Failed to open go source file: %s: %s", fileName, err)
			os.Exit(1)
		}

		comments := extractComments(parsedFile, fset, commentMarker)

		for _, comment := range comments {
			outer := enclosingBlock(fset, parsedFile, comment)
			for i, block := range profile.Blocks {
				if isWithin(&block, outer) && block.Count == 0 {
					profile.Blocks[i].Count = coverCount
				}
			}
		}

		fmt.Print(printProfile(profile))
	}
}

func printProfile(profile *cover.Profile) string {
	out := fmt.Sprintf("mode: %s\n", profile.Mode)
	for _, block := range profile.Blocks {
		if block.Count >= 0 {
			out += fmt.Sprintf("%s:%d.%d,%d.%d %d %d\n", profile.FileName, block.StartLine, block.StartCol, block.EndLine, block.EndCol, block.NumStmt, block.Count)
		}
	}
	return out
}

func newBlock(startLine int, startCol int, endLine int, endCol int) *cover.ProfileBlock {
	return &cover.ProfileBlock{StartLine: startLine, EndLine: endLine, StartCol: startCol, EndCol: endCol}
}

func isWithin(inner *cover.ProfileBlock, outer *cover.ProfileBlock) bool {
	return (inner.StartLine > outer.StartLine || (inner.StartLine == outer.StartLine && inner.StartCol >= outer.StartCol)) &&
		(inner.EndLine < outer.EndLine || (inner.EndLine == outer.EndLine && inner.EndCol <= outer.EndCol))
}

func enclosingBlock(fset *token.FileSet, root *ast.File, node ast.Node) *cover.ProfileBlock {
	//nocoverage
	parentNodes, _ := astutil.PathEnclosingInterval(root, node.Pos(), node.End())
	parentNode := parentNodes[0]
	start := fset.Position(parentNode.Pos())
	end := fset.Position(parentNode.End())
	return newBlock(start.Line, start.Column, end.Line, end.Column)
}

func extractComments(file *ast.File, fset *token.FileSet, text string) []*ast.Comment {
	//nocoverage
	var comments []*ast.Comment
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, text) {
				comments = append(comments, comment)
			}
		}
	}
	return comments
}
