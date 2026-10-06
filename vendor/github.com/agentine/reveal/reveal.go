package reveal

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Dump displays the passed parameters to standard output with type info,
// pointer addresses, and nested structure. It uses the default configuration.
func Dump(a ...interface{}) {
	Config.Fdump(os.Stdout, a...)
}

// Sdump returns a string with the passed parameters formatted using Dump style.
// It uses the default configuration.
func Sdump(a ...interface{}) string {
	var buf bytes.Buffer
	Config.Fdump(&buf, a...)
	return buf.String()
}

// Fdump formats and writes the passed parameters to the given io.Writer.
// It uses the default configuration.
func Fdump(w io.Writer, a ...interface{}) {
	Config.Fdump(w, a...)
}

// Sprint is a deep-aware replacement for fmt.Sprint.
func Sprint(a ...interface{}) string {
	return Config.Sprint(a...)
}

// Sprintf is a deep-aware replacement for fmt.Sprintf.
func Sprintf(format string, a ...interface{}) string {
	return Config.Sprintf(format, a...)
}

// Sprintln is a deep-aware replacement for fmt.Sprintln.
func Sprintln(a ...interface{}) string {
	return Config.Sprintln(a...)
}

// Fprint is a deep-aware replacement for fmt.Fprint.
func Fprint(w io.Writer, a ...interface{}) (int, error) {
	return Config.Fprint(w, a...)
}

// Fprintf is a deep-aware replacement for fmt.Fprintf.
func Fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	return Config.Fprintf(w, format, a...)
}

// Fprintln is a deep-aware replacement for fmt.Fprintln.
func Fprintln(w io.Writer, a ...interface{}) (int, error) {
	return Config.Fprintln(w, a...)
}

// Printf is a deep-aware replacement for fmt.Printf.
func Printf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stdout, format, wrapFormatters(a)...)
}

// Println is a deep-aware replacement for fmt.Println.
func Println(a ...interface{}) (int, error) {
	return fmt.Fprintln(os.Stdout, wrapFormatters(a)...)
}

// wrapFormatters wraps each argument with a reveal formatter.
func wrapFormatters(a []interface{}) []interface{} {
	args := make([]interface{}, len(a))
	for i, v := range a {
		args[i] = NewFormatter(v)
	}
	return args
}
