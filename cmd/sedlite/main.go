// every go file belongs to a package
// package main creates a runnable program
// the function main() inside package main is the entry point
package main

// import packages from the standard library
import (
	"fmt" // output text (stdout)
	"os"  // reads command line arguments

	"goshelltools/internal/sedlite" // import our sedlite package
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("example usage: ./sedlite -n 's/pattern/replacement/' [fileName](optional)")
		os.Exit(1) // exit with error code 1 (0 is success, anything else is error)
	}

	args := os.Args[1:] // skip program name

	noPrint := false

	// detect -n flag
	if args[0] == "-n" {
		noPrint = true
		args = args[1:] // remove -n from args
	}

	if len(args) < 1 {
		fmt.Println("error: missing sedlite command")
		os.Exit(1)
	}

	command := args[0]
	file := ""
	if len(args) >= 2 {
		file = args[1]
	}

	err := sedlite.Run(command, file, noPrint)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
