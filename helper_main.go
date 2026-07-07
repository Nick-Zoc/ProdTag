//go:build prodtaghelper

package main

import "os"

func main() {
	os.Exit(runHelper(os.Args[1:], os.Stdout, os.Stderr))
}
