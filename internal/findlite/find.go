package findlite

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type compareOp int

const (
	opExact   compareOp = iota // 0
	opLess                     // 1
	opGreater                  // 2
)

type query struct {
	namePattern   string    // name pattern for -name
	hasName       bool      // whether -name was specified
	ignoreCase    bool      // whether -iname was specified
	pathPattern   string    // path pattern for -path
	hasPath       bool      // whether -path was specified
	fileType      string    // file type for -type
	hasType       bool      // whether -type was specified
	emptyOnly     bool      // whether -empty was specified
	print         bool      // whether -print was specified
	explicitPrint bool      // whether -print was explicitly set
	doDelete      bool      // whether -delete was specified
	maxDepth      int       // max depth for -maxdepth
	hasMaxDepth   bool      // whether -maxdepth was specified
	minDepth      int       // min depth for -mindepth
	hasMinDepth   bool      // whether -mindepth was specified
	sizeOp        compareOp // size comparison operator
	sizeBytes     int64     // size in bytes
	hasSize       bool      // whether -size was specified
	mtimeOp       compareOp // mtime comparison operator
	mtimeDays     int       // days for mtime
	hasMtime      bool      // whether -mtime was specified
}

func Run(paths []string, expressions []string) error {
	q, err := parseExpressions(expressions)
	if err != nil {
		return err
	}

	for _, root := range paths {
		var toDelete []string

		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			// handle in Run to avoid computing depth unnecessarily
			depth, err := depthFromRoot(root, path)
			if err != nil {
				return err
			}
			if q.hasMaxDepth && depth > q.maxDepth {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}

			displayPath := formatDisplayPath(root, path)

			match, err := matches(q, displayPath, d, depth)
			if err != nil {
				return err
			}

			if match {
				if q.print {
					fmt.Println(displayPath)
				}
				if q.doDelete {
					toDelete = append(toDelete, path)
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		if q.doDelete {
			for i := len(toDelete) - 1; i >= 0; i-- {
				err := os.Remove(toDelete[i])
				if err != nil {
					return err
				}
			}
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
		case "-type":
			if i+1 >= len(exprs) {
				return query{}, errors.New("-type requires a value")
			}
			val := exprs[i+1]
			if val != "f" && val != "d" && val != "l" {
				return query{}, errors.New("invalid -type value (use f, d, or l)")
			}
			q.fileType = val
			q.hasType = true
			i++
		case "-empty":
			q.emptyOnly = true
		case "-maxdepth":
			if i+1 >= len(exprs) {
				return query{}, errors.New("-maxdepth requires a number")
			}
			val, err := strconv.Atoi(exprs[i+1])
			if err != nil || val < 0 {
				return query{}, errors.New("invalid -maxdepth value")
			}
			q.maxDepth = val
			q.hasMaxDepth = true
			i++
		case "-mindepth":
			if i+1 >= len(exprs) {
				return query{}, errors.New("-mindepth requires a number")
			}
			val, err := strconv.Atoi(exprs[i+1])
			if err != nil || val < 0 {
				return query{}, errors.New("invalid -mindepth value")
			}
			q.minDepth = val
			q.hasMinDepth = true
			i++
		case "-size":
			if i+1 >= len(exprs) {
				return query{}, errors.New("-size requires a value")
			}
			op, bytes, err := parseSize(exprs[i+1])
			if err != nil {
				return query{}, err
			}
			q.sizeOp = op
			q.sizeBytes = bytes
			q.hasSize = true
			i++
		case "-mtime":
			if i+1 >= len(exprs) {
				return query{}, errors.New("-mtime requires a value")
			}
			op, days, err := parseMtime(exprs[i+1])
			if err != nil {
				return query{}, err
			}
			q.mtimeOp = op
			q.mtimeDays = days
			q.hasMtime = true
			i++
		case "-print":
			q.print = true         // already default, redundant
			q.explicitPrint = true // mark as explicitly set
		case "-delete":
			q.doDelete = true
			if !q.explicitPrint {
				q.print = false
			}
		case "":
			return query{}, errors.New("empty expression")
		default:
			return query{}, fmt.Errorf("unsupported expression: %s", exprs[i])
		}
	}

	return q, nil
}

func formatDisplayPath(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	if rel == "." {
		return root
	}
	sep := string(filepath.Separator)
	// check if separator is already at the end of root
	if strings.HasSuffix(root, sep) {
		// return without separator added
		return root + rel
	}
	// return with separator added
	return root + sep + rel
}

func depthFromRoot(root string, path string) (int, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return 0, err
	}
	if rel == "." {
		return 0, nil
	}
	sep := string(filepath.Separator)
	return strings.Count(rel, sep) + 1, nil
}

func matches(q query, path string, d fs.DirEntry, depth int) (bool, error) {
	if q.hasMinDepth && depth < q.minDepth {
		return false, nil
	}

	if q.hasType {
		// want file, is not a file
		if q.fileType == "f" && !d.Type().IsRegular() {
			return false, nil
		}
		// want directory, is not directory
		if q.fileType == "d" && !d.IsDir() {
			return false, nil
		}
		// want symlink, is not symlink
		if q.fileType == "l" && d.Type()&os.ModeSymlink == 0 {
			return false, nil
		}
	}

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

	if q.emptyOnly {
		if d.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				return false, err
			}
			if len(entries) != 0 {
				return false, nil
			}
		} else {
			info, err := d.Info()
			if err != nil {
				return false, err
			}
			if info.Size() != 0 {
				return false, nil
			}
		}
	}

	if q.hasSize || q.hasMtime {
		info, err := d.Info()
		if err != nil {
			return false, err
		}
		if q.hasSize && !matchSize(q, info.Size()) {
			return false, nil
		}
		if q.hasMtime && !matchMtime(q, info.ModTime()) {
			return false, nil
		}
	}

	return true, nil
}

func parseSize(input string) (compareOp, int64, error) {
	op, rest := parseCompare(input)
	if rest == "" {
		return opExact, 0, errors.New("invalid -size value")
	}

	unit := byte(0)
	last := rest[len(rest)-1]
	if last < '0' || last > '9' {
		unit = last
		rest = rest[:len(rest)-1]
	}

	if rest == "" {
		return opExact, 0, errors.New("invalid -size value")
	}

	// parse the size number in base-10 up to 64 bits
	n, err := strconv.ParseInt(rest, 10, 64)
	if err != nil || n < 0 {
		return opExact, 0, errors.New("invalid -size value")
	}

	mult := int64(1)
	switch strings.ToLower(string(unit)) {
	case "":
		mult = 1
	case "c":
		mult = 1
	case "b":
		mult = 512
	case "k":
		mult = 1024
	case "m":
		mult = 1024 * 1024
	case "g":
		mult = 1024 * 1024 * 1024
	default:
		return opExact, 0, errors.New("invalid -size unit")
	}

	return op, n * mult, nil
}

func parseMtime(input string) (compareOp, int, error) {
	op, rest := parseCompare(input)
	if rest == "" {
		return opExact, 0, errors.New("invalid -mtime value")
	}
	n, err := strconv.Atoi(rest)
	if err != nil || n < 0 {
		return opExact, 0, errors.New("invalid -mtime value")
	}
	return op, n, nil
}

func parseCompare(input string) (compareOp, string) {
	if input == "" {
		return opExact, ""
	}
	switch input[0] {
	case '+':
		return opGreater, input[1:]
	case '-':
		return opLess, input[1:]
	default:
		return opExact, input
	}
}

func matchSize(q query, size int64) bool {
	switch q.sizeOp {
	case opGreater:
		return size > q.sizeBytes
	case opLess:
		return size < q.sizeBytes
	default:
		return size == q.sizeBytes
	}
}

func matchMtime(q query, modTime time.Time) bool {
	age := time.Since(modTime)
	if age < 0 {
		age = 0
	}
	days := int(age / (24 * time.Hour))

	switch q.mtimeOp {
	case opGreater:
		return days > q.mtimeDays
	case opLess:
		return days < q.mtimeDays
	default:
		return days == q.mtimeDays
	}
}
