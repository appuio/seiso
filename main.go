package main

import (
	"fmt"

	"github.com/appuio/seiso/cmd"
	log "github.com/sirupsen/logrus"
)

var (
	version = "unknown"
	commit  = "dirty"
	date    = "today"
)

func main() {
	cmd.SetVersion(fmt.Sprintf("%s, commit %s, date %s", version, commit, date))
	if err := cmd.Execute(); err != nil {
		log.WithError(err).Fatal("Command aborted.")
	}
}
