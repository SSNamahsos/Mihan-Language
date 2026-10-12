package main

import (
	"bufio"
	"fmt"
	"os"

	"mihan/internal/lsp"
	"mihan/internal/version"
)

func cmdLsp(args []string) error {

	out := bufio.NewWriter(os.Stdout)
	conn := lsp.NewConn(os.Stdin, out)
	lsp.NewServer(conn, version.Short())
	if err := conn.Serve(); err != nil {

		fmt.Fprintln(os.Stderr, "mihan lsp:", err)
		return err
	}
	return nil
}

const lspUsage = `mihan lsp — the Mihan language server

An editor starts this for you; you do not type it by hand. It answers what an
editor asks while someone is writing Mihan:

  what words could go here    the words of the language and the names in the file
  what is wrong with this     as the file is open, in Persian and in English
  what does this name mean    from the language's own tables, not a second list
  where was it written        a jump to the line that gave it a value
  what should this look like  the same tidying "mihan fmt" does

To watch what it is being asked, set mihan.trace.server in your editor to
"messages".
`
