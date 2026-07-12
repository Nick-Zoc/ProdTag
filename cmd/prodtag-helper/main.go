package main

import (
	"ProdTag/internal/core"
	"ProdTag/internal/helpercli"
	"os"
)

func main() { os.Exit(helpercli.Run(os.Args[1:], os.Stdout, os.Stderr, core.StartPlayback)) }
