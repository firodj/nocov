package main

import (
	"path/filepath"
	"regexp"
)

// Taken from: https://github.com/boumenot/gocover-cobertura/blob/master/ignore.go
// As golint-ci referencing https://golang.org/s/generatedcode, be laxer
var genCodeRe = regexp.MustCompile(`(?im)^//.*(?:code generated|do not edit|autogenerated file)`)

type Ignore struct {
	Dirs           *regexp.Regexp
	Files          *regexp.Regexp
	GeneratedFiles bool
	cache          map[string]bool
}

func (i *Ignore) Match(fileName string, data []byte) (ret bool) {
	if i.cache == nil {
		i.cache = map[string]bool{}
	} else if match, exists := i.cache[fileName]; exists {
		return match
	}

	dir := filepath.Dir(fileName)

	if i.dirMatch(dir) ||
		(i.Files != nil && i.Files.MatchString(fileName)) {
		ret = true
	} else if i.GeneratedFiles {
		if data == nil {
			return false // no cache if no content provided
		}

		if len(data) > 256 {
			data = data[:256]
		}
		ret = genCodeRe.Match(data)
	}

	i.cache[fileName] = ret

	return ret
}

func (i *Ignore) dirMatch(dir string) bool {
	if i.Dirs == nil {
		return false
	}

	for {
		if i.Dirs.MatchString(dir) {
			return true
		}
		dir, _ = filepath.Split(dir)
		if dir == "" {
			return false
		}
		dir = dir[:len(dir)-1] // without last separator
	}
}
