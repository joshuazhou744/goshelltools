package main

import (
	"fmt"
	"os"
	"strings"

	"goshelltools/internal/findlite" // import our findlite package
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("example usage: ./findlite [path...] [expression...]")
		os.Exit(1) // exit with error code 1 (0 is success, anything else is error)
	}

	args := os.Args[1:] // skip program name
	paths, expressions := splitPathsAndExpressions(args)

	if len(paths) == 0 {
		paths = []string{"."}
	}

	err := findlite.Run(paths, expressions)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func splitPathsAndExpressions(args []string) ([]string, []string) {
	for i, arg := range args {
		if isExpressionStart(arg) {
			return args[:i], args[i:]
		}
	}
	return args, nil
}

func isExpressionStart(arg string) bool {
	if arg == "!" || arg == "(" || arg == ")" {
		return true
	}
	return strings.HasPrefix(arg, "-")
}
