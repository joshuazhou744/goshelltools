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
