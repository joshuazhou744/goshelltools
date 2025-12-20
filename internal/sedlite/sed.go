package sedlite // package name

import (
	"bufio"   // provides buffered I/O, will use to read files line by line
	"errors"  // error handling
	"fmt"     // formatted I/O
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
	CommandPrintAll                      // 3
	CommandQuit                          // 4
)

// structure for one fully parsed sed command
type Command struct {
	Type        CommandType
	Regex       *regexp.Regexp
	Replacement string // only used for substitution
}

func Run(commands []string, files []string, noPrint bool) error {
	cmds, err := parseCommands(commands)

	// if error, return it
	if err != nil {
		return err
	}

	if len(files) == 0 {
		_, err := runOnReader(cmds, os.Stdin, noPrint)
		return err
	}

	for _, file := range files {
		f, err := os.Open(file) // open the file
		if err != nil {
			return err
		}

		quit, runErr := runOnReader(cmds, f, noPrint)
		if closeErr := f.Close(); closeErr != nil && runErr == nil {
			runErr = closeErr
		}
		if runErr != nil {
			return runErr
		}
		if quit {
			return nil
		}
	}

	return nil
}

func runOnReader(cmds []*Command, reader io.Reader, noPrint bool) (bool, error) {
	// read the reader line by line using a budio.Scanner
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text() // get a line of text from the scanner
		printLine := !noPrint  // by default we print the line unless noPrint is true
		skipRemaining := false

		for _, cmd := range cmds {
			if skipRemaining {
				break
			}

			switch cmd.Type {
			case CommandSubstitute:
				line = cmd.Regex.ReplaceAllString(line, cmd.Replacement) // apply regex substitution
			case CommandDelete:
				if cmd.Regex.MatchString(line) {
					printLine = false // do not print the line if it matches the regex
					skipRemaining = true
				}
			case CommandPrint:
				if cmd.Regex.MatchString(line) {
					fmt.Println(line) // print the line if regex is found in line (matching)
				}
				printLine = false // prevent default print (only print if matched)
			case CommandQuit:
				if cmd.Regex.MatchString(line) {
					if printLine {
						fmt.Println(line)
					}
					return true, nil // exit the program if matched
				}
			case CommandPrintAll:
				fmt.Println(line) // explicitly print the line
			}
		}

		if printLine {
			fmt.Println(line) // print the line if printLine is true
		}
	}

	return false, scanner.Err()
}

func parseCommands(inputs []string) ([]*Command, error) {
	if len(inputs) == 0 {
		return nil, errors.New("missing sedlite command")
	}

	cmds := make([]*Command, 0, len(inputs))
	for _, input := range inputs {
		cmd, err := parseCommand(input)
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

// parse command helper function (match the command type and call the appropriate parser)
func parseCommand(input string) (*Command, error) {
	if strings.HasPrefix(input, "s/") {
		return parseSubstitute(input)
	} else if strings.HasPrefix(input, "/") {
		if strings.HasSuffix(input, "/d") {
			return parseDelete(input)
		} else if strings.HasSuffix(input, "/p") {
			return parsePrint(input)
		} else if strings.HasSuffix(input, "/q") {
			return parseQuit(input)
		}
	} else if input == "p" {
		return printAll()
	}

	return nil, errors.New("unsupported command")
}

// parse delete command
func parseDelete(input string) (*Command, error) {
	if len(input) < 4 || !strings.HasPrefix(input, "/") || !strings.HasSuffix(input, "/d") {
		return nil, errors.New("invalid delete syntax")
	}

	pattern := input[1 : len(input)-2]

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &Command{
		Type:  CommandDelete,
		Regex: re,
	}, nil
}

// parse print matching command
func parsePrint(input string) (*Command, error) {
	if len(input) < 4 || !strings.HasPrefix(input, "/") || !strings.HasSuffix(input, "/p") {
		return nil, errors.New("invalid print syntax")
	}

	pattern := input[1 : len(input)-2]

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &Command{
		Type:  CommandPrint,
		Regex: re,
	}, nil
}

// parste quit command
func parseQuit(input string) (*Command, error) {
	if len(input) < 4 || !strings.HasPrefix(input, "/") || !strings.HasSuffix(input, "/q") {
		return nil, errors.New("invalid quit syntax")
	}

	re, err := regexp.Compile(input[1 : len(input)-2])
	if err != nil {
		return nil, err
	}

	return &Command{
		Type:  CommandQuit,
		Regex: re,
	}, nil
}

// print all command
func printAll() (*Command, error) {
	return &Command{
		Type: CommandPrintAll,
	}, nil
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
