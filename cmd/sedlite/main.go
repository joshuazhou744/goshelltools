// every go file belongs to a package
// package main creates a runnable program
// the function main() inside package main is the entry point
package main

// import packages from the standard library
import (
	"fmt" // output text (stdout)
	"os"  // reads command line arguments

	"github.com/joshuazhou744/goshelltools.git/internal/sedlite"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: sedlite 's/pattern/replacement/' [file](optional)")
		os.Exit(1) // exit with error code 1 (0 is success, anything else is error)
	}

	command := os.Args[1] // sedlite
	err := sedlite.Run(command, "")
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
