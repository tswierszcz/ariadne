// Command ariadne reports which known vulnerabilities (CVE/GHSA) in a JVM
// application's dependencies are reachable from the application's own code.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("ariadne", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		fmt.Fprintf(stdout, "ariadne %s\n", version())
		return 0
	}

	fmt.Fprintln(stderr, "ariadne: analysis is not implemented yet (see PLAN.md)")
	return 2
}

// version returns the module version, or the VCS revision for local builds.
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	rev, dirty := "", false
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return "devel"
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if dirty {
		rev += "-dirty"
	}
	return "devel+" + rev
}
