## goshelltools

Building some unix tools to help me learn Go.

## Project structure

```
goshelltools/
├── cmd/
│   └── sedlite/
│       └── main.go
├── internal/
│   └── sedlite/
│       └── sed.go
└── go.mod
```

- `cmd/` contains executables
    - `main.go` in each tools commands is the entry point where the program starts running
- `internal/` contains code used by the program which is private (hence name internal)
    - Keeps logic separate from the startup code, like destructured services


## `sedlite`

### Commands
- Substitute: `s/pattern/replacement/`
- Delete entire line of the pattern (all instances): `/pattern/d`
- Print matching: `/pattern/p`
- Quit after pattern is found: `/pattern/q`
- Print all lines: `p`

### Flags
- `-n`: Disable default printing (no-print flag)

## Build and run the executable

Build
```bash
go build -o toolname /path/to/main.go
```

Run
```bash
./sedlite [command] [fileName](optional)
```