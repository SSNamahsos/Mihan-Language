package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"mihan/internal/version"
)

const usage = `mihan — a programming language that reads the way Persian is spoken

  mihan run <file.mhn>     run a program
  mihan run .              run the folder, using its main.mhn
  mihan run --open <file.mhn>
                           run a site and open it in the default browser
  mihan serve <file.mhn>   the same as run --open
  mihan repl               start the line-at-a-time reader
  mihan build <file.mhn>   make one executable that carries the program
                           (--os linux|windows|android, --arch amd64|arm64, --window, --open-browser, -o name)
  mihan preview <file.mhn> --android
                           show it as its Android app would, in a phone frame
  mihan check <file.mhn>   look for mistakes without running anything
  mihan compile <file.mhn> show the instructions a program becomes
  mihan test               run every documented example as a test
  mihan doc               the whole language reference, printed
  mihan fmt <file.mhn>     put a file in Mihan's usual shape
  mihan new <name>         start a new program in a folder
  mihan install            put Mihan where an editor finds it
  mihan lsp                the language server, which an editor starts by itself
  mihan env                say where everything is
  mihan version            say which build this is
  mihan help               this list

A Mihan program is a script: whatever it says at the top is the program. There
is no main to write.

  mihan run salam.mhn

To start using Mihan, run "mihan install" once. It puts the compiler and the
language server where your editor can find them, exactly as "go install" does.
`

func main() {

	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(400)
	}

	if mainFile, ok, err := embeddedProgram(); ok {
		if err == nil {
			err = cmdRun(append([]string{mainFile}, os.Args[1:]...))
		}
		if err != nil {
			if msg := err.Error(); msg != "" {
				fmt.Fprintf(os.Stderr, "%v\n", msg)
			}
			os.Exit(1)
		}
		return
	}
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprint(os.Stdout, usage)
		return
	}
	cmd, rest := args[0], args[1:]
	var err error
	switch cmd {
	case "run", "start":
		err = cmdRun(rest)
	case "serve", "s":

		err = cmdRun(append([]string{"--open"}, rest...))
	case "repl":
		err = cmdRepl(rest)
	case "build", "b":
		err = cmdBuild(rest)
	case "preview", "p":
		err = cmdPreview(rest)
	case "check", "chk":
		err = cmdCheck(rest)
	case "compile", "asr":
		err = cmdCompile(rest)
	case "test", "t":
		err = cmdTest(rest)
	case "fmt", "f":
		err = cmdFmt(rest)
	case "new":
		err = cmdNew(rest)
	case "lsp", "language-server":
		if len(rest) > 0 && (rest[0] == "-h" || rest[0] == "--help") {
			fmt.Print(lspUsage)
			return
		}
		err = cmdLsp(rest)
	case "install":
		err = cmdInstall(rest)
	case "env":
		err = cmdEnv(rest)
	case "doc", "docs", "reference":
		err = cmdDoc(rest)
	case "version", "v", "--version":
		fmt.Println(version.String())
	case "help", "h", "--help", "-h":
		fmt.Fprint(os.Stdout, usage)
	default:

		if strings.ContainsAny(cmd, `/\.`) || driver_Exists(cmd) {
			err = cmdRun(args)
			break
		}
		fmt.Fprintf(os.Stderr, "mihan: I do not know the command %q\n  try `mihan help`\n", cmd)
		os.Exit(2)
	}
	if err != nil {

		if msg := err.Error(); msg != "" {
			fmt.Fprintf(os.Stderr, "mihan: %v\n", msg)
		}
		os.Exit(1)
	}
}

func driver_Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fatal(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
