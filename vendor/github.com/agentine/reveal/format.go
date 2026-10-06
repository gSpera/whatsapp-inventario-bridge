package reveal

import (
	"bytes"
	"fmt"
	"io"
)

// formatter implements fmt.Formatter for deep printing Go values.
// It wraps a value and a ConfigState, allowing use with fmt.Printf and friends.
type formatter struct {
	value interface{}
	cs    *ConfigState
}

// NewFormatter returns a fmt.Formatter that formats the given value using reveal.
// The returned Formatter supports the following verbs:
//
//	%v   — inline format (same as fmt %v but with deep traversal)
//	%+v  — inline format with field names (for structs)
//	%#v  — Go-syntax representation
//	%d   — dump format (same as Dump output)
//
// Use it with fmt.Sprintf, fmt.Fprintf, etc:
//
//	fmt.Sprintf("%v", reveal.NewFormatter(myStruct))
func NewFormatter(v interface{}) fmt.Formatter {
	return &formatter{value: v, cs: &Config}
}

// Format implements fmt.Formatter.
func (f *formatter) Format(s fmt.State, verb rune) {
	switch verb {
	case 'd':
		// Dump format.
		var buf bytes.Buffer
		d := dumpState{
			w:      &buf,
			cs:     f.cs,
			seen:   make(map[uintptr]bool),
			depth:  0,
			indent: "",
		}
		d.dump(f.value)
		io.WriteString(s, buf.String())

	case 'v':
		if s.Flag('#') {
			// Go-syntax representation.
			fmt.Fprintf(s, "%#v", f.value)
		} else if s.Flag('+') {
			// Verbose with field names — use dump format.
			var buf bytes.Buffer
			d := dumpState{
				w:      &buf,
				cs:     f.cs,
				seen:   make(map[uintptr]bool),
				depth:  0,
				indent: "",
			}
			d.dump(f.value)
			io.WriteString(s, buf.String())
		} else {
			// Standard value format.
			fmt.Fprintf(s, "%v", f.value)
		}

	case 's':
		// String format — use dump.
		var buf bytes.Buffer
		d := dumpState{
			w:      &buf,
			cs:     f.cs,
			seen:   make(map[uintptr]bool),
			depth:  0,
			indent: "",
		}
		d.dump(f.value)
		io.WriteString(s, buf.String())

	default:
		// For all other verbs, fall back to fmt default formatting.
		fmt.Fprintf(s, fmt.FormatString(s, verb), f.value)
	}
}

// newConfigFormatter returns a formatter bound to a specific ConfigState.
func newConfigFormatter(cs *ConfigState, v interface{}) fmt.Formatter {
	return &formatter{value: v, cs: cs}
}

// NewFormatter returns a fmt.Formatter bound to this ConfigState.
func (c *ConfigState) NewFormatter(v interface{}) fmt.Formatter {
	return newConfigFormatter(c, v)
}

// Sprint is a deep-aware replacement for fmt.Sprint.
// Each argument is formatted using reveal's dump traversal.
func (c *ConfigState) Sprint(a ...interface{}) string {
	return c.dumpArgs(a)
}

// Sprintf is a deep-aware replacement for fmt.Sprintf.
// Arguments referenced by format verbs are formatted using reveal's deep traversal.
func (c *ConfigState) Sprintf(format string, a ...interface{}) string {
	args := make([]interface{}, len(a))
	for i, v := range a {
		args[i] = newConfigFormatter(c, v)
	}
	return fmt.Sprintf(format, args...)
}

// Sprintln is a deep-aware replacement for fmt.Sprintln.
func (c *ConfigState) Sprintln(a ...interface{}) string {
	return c.dumpArgs(a) + "\n"
}

// Fprint is a deep-aware replacement for fmt.Fprint.
func (c *ConfigState) Fprint(w io.Writer, a ...interface{}) (int, error) {
	return io.WriteString(w, c.dumpArgs(a))
}

// Fprintf is a deep-aware replacement for fmt.Fprintf.
func (c *ConfigState) Fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	args := make([]interface{}, len(a))
	for i, v := range a {
		args[i] = newConfigFormatter(c, v)
	}
	return fmt.Fprintf(w, format, args...)
}

// Fprintln is a deep-aware replacement for fmt.Fprintln.
func (c *ConfigState) Fprintln(w io.Writer, a ...interface{}) (int, error) {
	return io.WriteString(w, c.dumpArgs(a)+"\n")
}

// dumpArgs formats multiple arguments as space-separated dump strings.
func (c *ConfigState) dumpArgs(a []interface{}) string {
	parts := make([]string, len(a))
	for i, v := range a {
		var buf bytes.Buffer
		d := dumpState{
			w:      &buf,
			cs:     c,
			seen:   make(map[uintptr]bool),
			depth:  0,
			indent: "",
		}
		d.dump(v)
		parts[i] = buf.String()
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}
