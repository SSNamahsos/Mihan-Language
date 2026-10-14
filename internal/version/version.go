package version

import (
	"fmt"
	"runtime"
	"strings"
)

const Name = "Mihan"

const Version = "0.7.8"

const Stage = "stage 1 — scripts, functions, classes and the standard library"

const Codename = "salam"

func String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s (%s)\n", Name, Version, Codename)
	fmt.Fprintf(&b, "  %s\n", Stage)
	fmt.Fprintf(&b, "  %s on %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return b.String()
}

func Short() string { return Name + " " + Version }
