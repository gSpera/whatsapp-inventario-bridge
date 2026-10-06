# reveal

[![Go Reference](https://pkg.go.dev/badge/github.com/agentine/reveal.svg)](https://pkg.go.dev/github.com/agentine/reveal)
[![Go 1.21+](https://img.shields.io/badge/go-1.21%2B-blue)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

A modern, maintained, and bug-fixed deep pretty printer for Go data structures. Drop-in replacement for [go-spew](https://github.com/davecgh/go-spew).

## Installation

```bash
go get github.com/agentine/reveal
```

## Quick Start

```go
import "github.com/agentine/reveal"

reveal.Dump(myStruct)          // print to stdout
s := reveal.Sdump(myStruct)    // return as string
reveal.Fdump(w, myStruct)      // print to io.Writer
```

Example output:

```
(main.Person) {
 Name: (string) (len=5) "Alice",
 Age: (int) 30,
 Addr: (*main.Address)(0xc0000b4000)(
  (main.Address) {
   City: (string) (len=10) "Wonderland"
  })
}
```

## Migration from go-spew

Change one import — nothing else needs to change:

```go
// Before
import "github.com/davecgh/go-spew/spew"
spew.Dump(myStruct)

// After
import "github.com/agentine/reveal"
reveal.Dump(myStruct)
```

Output format is identical by default. All go-spew API functions (`Dump`, `Sdump`, `Fdump`, `Sprint`, `Sprintf`, `Sprintln`, `Fprint`, `Fprintf`, `Fprintln`, `NewFormatter`) are supported, with the same signatures.

## API

### Package-level functions

All functions use the default `Config`. Each function is also available as a method on `ConfigState` for custom configurations.

| Function | Description |
|---|---|
| `Dump(a ...)` | Print to stdout |
| `Sdump(a ...) string` | Return as string |
| `Fdump(w, a ...)` | Print to io.Writer |
| `Sprint(a ...) string` | Deep-aware fmt.Sprint |
| `Sprintf(format, a ...) string` | Deep-aware fmt.Sprintf |
| `Sprintln(a ...) string` | Deep-aware fmt.Sprintln |
| `Fprint(w, a ...) (int, error)` | Deep-aware fmt.Fprint |
| `Fprintf(w, format, a ...) (int, error)` | Deep-aware fmt.Fprintf |
| `Fprintln(w, a ...) (int, error)` | Deep-aware fmt.Fprintln |
| `Printf(format, a ...) (int, error)` | Deep-aware fmt.Printf |
| `Println(a ...) (int, error)` | Deep-aware fmt.Println |
| `NewFormatter(v) fmt.Formatter` | Wrap value for use with fmt verbs |

### fmt verb support

Values wrapped with `NewFormatter` support these verbs:

| Verb | Output |
|---|---|
| `%v` | Standard fmt value format |
| `%+v` | Dump format with field names |
| `%#v` | Go-syntax representation |
| `%d` | Dump format (same as `Dump` output) |
| `%s` | Dump format as string |

### Configuration

Modify the global default:

```go
reveal.Config.MaxDepth = 5
reveal.Config.SortKeys = true
reveal.Config.Dump(myStruct)
```

Or create an isolated configuration:

```go
c := reveal.ConfigState{
    Indent:   "    ",
    SortKeys: true,
    MaxDepth: 3,
}
c.Dump(myStruct)
```

| Field | Default | Description |
|---|---|---|
| `Indent` | `" "` | Indentation string per level |
| `MaxDepth` | `0` | Max recursion depth (0 = unlimited) |
| `DisableMethods` | `false` | Skip Stringer/Error methods |
| `DisablePointerMethods` | `false` | Skip pointer receiver methods |
| `DisablePointerAddresses` | `false` | Hide pointer addresses |
| `DisableCapacities` | `false` | Hide slice capacities |
| `ContinueOnMethod` | `false` | Show internals after method output |
| `SortKeys` | `false` | Sort map keys for deterministic output |
| `SpewKeys` | `false` | Use dump format for map keys |
| `MaxSize` | `10485760` (10 MB) | Max output bytes (0 = unlimited) |
| `RecoverPanics` | `true` | Recover from Stringer/Error panics |
| `OmitNilPointers` | `false` | Suppress nil pointer struct fields |
| `OmitUnexported` | `false` | Suppress unexported struct fields |
| `HexIntegers` | `false` | Format integers as hexadecimal |

## Safety Features

**`MaxSize`** (default 10 MB): Output is truncated and a notice appended once the byte limit is reached. Prevents memory exhaustion when dumping deeply nested or self-referential structures.

**`RecoverPanics`** (default `true`): All `Stringer` and `error` method calls are wrapped in a deferred recover. If a method panics, reveal falls back to raw value output and includes `<panic in method: ...>` in the output rather than crashing the caller. Set to `false` to let panics propagate (useful in tests).

## Bug Fixes vs go-spew

| Issue | Description | Fix |
|---|---|---|
| [#145](https://github.com/davecgh/go-spew/issues/145) | Memory exhaustion on deeply nested structures | `MaxSize` output limit (default 10 MB) |
| [#141](https://github.com/davecgh/go-spew/issues/141) | Panic with custom Stringer on map values | `RecoverPanics` safe method invocation |
| [#115](https://github.com/davecgh/go-spew/issues/115) | Panic with custom Stringer on maps | `RecoverPanics` safe method invocation |
| [#144](https://github.com/davecgh/go-spew/issues/144) | Panic with wrapped custom errors | Panic recovery covers `error.Error()` calls |
| [#108](https://github.com/davecgh/go-spew/issues/108) | Panic sorting maps with unexported field keys | Safe key comparison logic |
| — | Stringer recursion loop | Detected and broken automatically (pointer and type-depth tracking) |

## Build Tags

By default, reveal uses `unsafe` reflection to access unexported struct fields (same technique as go-spew). To disable this:

```bash
go build -tags safe ./...
# or
go build -tags disableunsafe ./...
```

With either tag, unexported fields are shown as `<unexported>` instead of their values.

## License

MIT
