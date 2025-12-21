// every go file belongs to a package
// package main creates a runnable program
// the function main() inside package main is the entry point
package main

// import packages from the st	andard library
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

	commands := []string{}
	files := []string{}

	if args[0] == "-e" {
		for i := 0; i < len(args); {
			if args[i] == "-e" {
				if i+1 >= len(args) {
					fmt.Println("error: -e flag requires a command")
					os.Exit(1)
				}
				commands = append(commands, args[i+1])
				i += 2
				continue
			}

			files = append(files, args[i]) // add file
			i++
		}
	} else {
		commands = append(commands, args[0])
		if len(args) >= 2 {
			files = append(files, args[1:]...) // append each remaining arg as a file
		}
	}

	if len(commands) == 0 {
		fmt.Println("error: missing sedlite command")
		os.Exit(1)
	}

	err := sedlite.Run(commands, files, noPrint)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
