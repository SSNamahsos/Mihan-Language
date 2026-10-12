package main

import (
	"fmt"
	"os"

	"mihan/internal/lsp"
	"mihan/internal/version"
)

func main() {
	for _, a := range os.Args[1:] {
		switch a {
		case "-v", "--version", "version":
			fmt.Println("mihan-ls", version.String())
			return
		case "-h", "--help", "help":
			fmt.Print(usage)
			return
		}
	}

	conn := lsp.NewConn(os.Stdin, os.Stdout)
	lsp.NewServer(conn, version.Short())
	if err := conn.Serve(); err != nil {
		fmt.Fprintln(os.Stderr, "mihan-ls:", err)
		os.Exit(1)
	}
}

const usage = `mihan-ls — the Mihan language server

An editor starts this program; you do not type it by hand. It answers the
questions an editor asks while someone is writing Mihan: what words could go
here, what is wrong with this file, what does this name mean, where was it
written, and what should this file look like tidied.

  mihan install        put mihan and mihan-ls where the editor can find them
`
