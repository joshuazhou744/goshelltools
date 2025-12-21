package findlite

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type query struct {
	namePattern string // name pattern for -name
	hasName     bool   // whether -name was specified
	ignoreCase  bool   // whether -iname was specified
	pathPattern string // path pattern for -path
	hasPath     bool   // whether -path was specified
	print       bool   // whether -print was specified
}

func Run(paths []string, expressions []string) error {
	q, err := parseExpressions(expressions)
	if err != nil {
		return err
	}

	for _, root := range paths {
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			displayPath := path
			if root == "." && path != "." {
				displayPath = "." + string(filepath.Separator) + path
			}

			match, err := matches(q, displayPath, d)
			if err != nil {
				return err
			}

			if match && q.print {
				fmt.Println(displayPath)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func parseExpressions(exprs []string) (query, error) {
	q := query{print: true}

	for i := 0; i < len(exprs); i++ {
		switch exprs[i] {
		case "-name":
			if i+1 >= len(exprs) {
				return query{}, errors.New("-name requires a pattern")
			}
			q.namePattern = exprs[i+1]
			q.hasName = true
			i++
		case "-iname":
			if i+1 >= len(exprs) {
				return query{}, errors.New("-iname requires a pattern")
			}
			q.namePattern = exprs[i+1]
			q.hasName = true
			q.ignoreCase = true
			i++
		case "-path":
			if i+1 >= len(exprs) {
				return query{}, errors.New("-path requires a pattern")
			}
			q.pathPattern = exprs[i+1]
			q.hasPath = true
			i++
		case "-print":
			q.print = true // already default, redundant but explicit
		case "":
			return query{}, errors.New("empty expression")
		default:
			return query{}, fmt.Errorf("unsupported expression: %s", exprs[i])
		}
	}

	return q, nil
}

func matches(q query, path string, d fs.DirEntry) (bool, error) {
	if q.hasName {
		pattern := q.namePattern
		name := d.Name()
		if q.ignoreCase {
			pattern = strings.ToLower(pattern)
			name = strings.ToLower(name)
		}
		ok, err := filepath.Match(pattern, name)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	if q.hasPath {
		ok, err := filepath.Match(q.pathPattern, path)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}
