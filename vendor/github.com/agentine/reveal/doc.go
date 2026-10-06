// Package reveal is a deep pretty printer for Go data structures.
//
// It is a modern, maintained, and bug-fixed replacement for github.com/davecgh/go-spew.
// The API is fully compatible with go-spew — change the import path and everything works.
//
// Basic usage:
//
//	reveal.Dump(myStruct)          // print to stdout
//	s := reveal.Sdump(myStruct)    // return as string
//	reveal.Fdump(w, myStruct)      // print to io.Writer
//
// Configuration:
//
//	config := reveal.Config
//	config.MaxDepth = 5
//	config.Dump(myStruct)
package reveal
