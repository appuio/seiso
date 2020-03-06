package main

import (
	"fmt"
	"github.com/appuio/image-cleanup/cmd"
)

var (
	version = "unknown"
	commit  = "dirty"
	date    = "today"
)

func main() {
	cmd.SetVersion(fmt.Sprintf("%s, commit %s, date %s", version, commit, date))
	cmd.Execute()
}
