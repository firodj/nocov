package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/cover"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type Source struct {
	Path string
}

func usage() {
	//nocoverage
	fmt.Fprintf(os.Stderr, "nocov - a golang tool to remove blocks of code from the coverage report\n\n")
	flag.PrintDefaults()
}

func getPackageName(filename string) string {
	pkgName, _ := filepath.Split(filename)
	// TODO(boumenot): Windows vs. Linux
	return strings.TrimRight(strings.TrimRight(pkgName, "\\"), "/")
}

func getPackages(profiles []*cover.Profile) ([]*packages.Package, error) {
	var pkgNames []string
	for _, profile := range profiles {
		pkgNames = append(pkgNames, getPackageName(profile.FileName))
	}
	return packages.Load(&packages.Config{Mode: packages.NeedFiles | packages.NeedModule}, pkgNames...)
}

func appendIfUnique(sources []*Source, dir string) []*Source {
	for _, source := range sources {
		if source.Path == dir {
			return sources
		}
	}
	return append(sources, &Source{dir})
}

func findAbsFilePath(pkg *packages.Package, profileName string) string {
	filename := filepath.Base(profileName)
	for _, fullpath := range pkg.GoFiles {
		if filepath.Base(fullpath) == filename {
			return fullpath
		}
	}
	return ""
}

func main() {

	var coverFilename string
	var commentMarker string
	var coverCount int
	flag.StringVar(&coverFilename, "coverprofile", "c.out", "The filename of the cover profile as generated by : go test -coverprofile=<coverprofile>")
	flag.StringVar(&commentMarker, "commentMarker", "//nocoverage", "The comment to search for in the code that is used to define a block that cannot be covered by the tests")
	flag.IntVar(&coverCount, "coverCount", -1, "Blocks marked with the <commentMarker> will be assigned to have been covered <coverCount> times. If this number is negative, the block is removed from the coverage report")

	var ignore Ignore

	flag.BoolVar(&ignore.GeneratedFiles, "ignore-gen-files", false, "ignore generated files")
	ignoreDirsRe := flag.String("ignore-dirs", "", "ignore dirs matching this regexp")
	ignoreFilesRe := flag.String("ignore-files", "", "ignore files matching this regexp")

	flag.Usage = usage
	flag.Parse()

	profiles, err := cover.ParseProfiles(coverFilename)
	if err != nil {
		//nocoverage
		log.Fatalf("Failed to open cover profile file: %s: %s", coverFilename, err)
		os.Exit(1)
	}

	if *ignoreDirsRe != "" {
		ignore.Dirs, err = regexp.Compile(*ignoreDirsRe)
		if err != nil {
			log.Fatalf("Bad -ignore-dirs regexp: %s\n", err)
			os.Exit(1)
		}
	}

	if *ignoreFilesRe != "" {
		ignore.Files, err = regexp.Compile(*ignoreFilesRe)
		if err != nil {
			log.Fatalf("Bad -ignore-files regexp: %s\n", err)
			os.Exit(1)
		}
	}

	pkgs, err := getPackages(profiles)
	if err != nil {
		log.Fatalf("Error load packages: %s", err)
		os.Exit(1)
	}

	sources := make([]*Source, 0)
	pkgMap := make(map[string]*packages.Package)
	for _, pkg := range pkgs {
		sources = appendIfUnique(sources, pkg.Module.Dir)
		pkgMap[pkg.ID] = pkg
	}

	modeSet := true

	for _, profile := range profiles {
		pkgName := getPackageName(profile.FileName)
		pkgPkg := pkgMap[pkgName]
		if pkgPkg == nil || pkgPkg.Module == nil {
			log.Fatalf("package required when using go modules")
			os.Exit(1)
		}

		fileName := profile.FileName[len(pkgPkg.Module.Path)+1:]
		absFilePath := findAbsFilePath(pkgPkg, profile.FileName)

		//nocoverage

		fset := token.NewFileSet()
		parsedFile, err := parser.ParseFile(fset, absFilePath, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("Failed to open go source file: %s: %s", absFilePath, err)
			os.Exit(1)
		}

		data, err := ioutil.ReadFile(absFilePath)
		if err != nil {
			log.Fatalf("Failed to read go source file: %s: %s", absFilePath, err)
			os.Exit(1)
		}

		if ignore.Match(fileName, data) {
			log.Printf("Cov ignoring: %s", fileName)
			continue
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

		if modeSet {
			out := fmt.Sprintf("mode: %s\n", profile.Mode)
			fmt.Print(out)
			modeSet = false
		}
		fmt.Print(printProfile(profile))
	}
}

func printProfile(profile *cover.Profile) string {
	var out string
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
