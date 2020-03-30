package main

import (
	"fmt"
	"os"

	"github.com/appuio/seiso/cmd"
)

var (
	version = "unknown"
	commit  = "dirty"
	date    = "today"
)

func main() {
	cmd.SetVersion(fmt.Sprintf("%s, commit %s, date %s", version, commit, date))
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
