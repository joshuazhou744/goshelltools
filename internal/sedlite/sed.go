package sedlite // package name

import (
	"bufio"  // provides buffered I/O, will use to read files line by line
	"errors" // error handling
	"fmt"
	"io"      // basic I/O interfaces (Reader, Writer)
	"os"      // access to operating system functionality (files, stdin, cmd arguments)
	"regexp"  // regular expression matching and text replacement
	"strings" // string manipulation functions
)

// CommandType represents the type of sed command with an integer
type CommandType int

// possible CommandType values
const (
	CommandSubstitute CommandType = iota // iota starts at 0 and increments by 1 for each constant
	CommandDelete                        // 1
	CommandPrint                         // 2
)

// structure for one fully parsed sed command
type Command struct {
	Type        CommandType
	Regex       *regexp.Regexp
	Replacement string // only used for substitution
}

func Run(command string, file string) error {
	cmd, err := parseCommand(command)

	// if error, return it
	if err != nil {
		return err
	}

	// defines the reader as a variable of type io.Reader (comes after the name in go)
	// for example we can use os.Stdin which reads from standard input (command line)
	var reader io.Reader = os.Stdin

	// read the reader line by line using a budio.Scanner
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text() // get a line of text from the scanner

		switch cmd.Type {
		case CommandSubstitute:
			out := cmd.Regex.ReplaceAllString(line, cmd.Replacement) // apply regex substitution
			fmt.Println(out)                                         // print the modified line
		}
	}

	return scanner.Err()
}

// parse command helper function (match the command type and call the appropriate parser)
func parseCommand(input string) (*Command, error) {
	if strings.HasPrefix(input, "s/") {
		return parseSubstitute(input)
	}

	return nil, errors.New("unsupported command")
}

// parse substitute command
func parseSubstitute(input string) (*Command, error) {
	body := input[2:]                 // remove "s/" prefix
	parts := strings.Split(body, "/") // split by "/"

	if len(parts) < 2 {
		return nil, errors.New("invalid substitution syntax")
	}

	pattern := parts[0]     // first part is the pattern
	replacement := parts[1] // second part is the replacement

	re, err := regexp.Compile(pattern) // compile the regex pattern
	// return if error
	if err != nil {
		return nil, err
	}

	// return the Command struct and a nil error (no error)
	return &Command{
		Type:        CommandSubstitute,
		Regex:       re,
		Replacement: replacement,
	}, nil
}
