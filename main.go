package main

import (
	"log"

	logs "github.com/appscode/go/log/golog"
	"github.com/k8sdb/operator/cmds"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()
	if err := cmds.NewRootCmd("mongoExporter").Execute(); err != nil {
		log.Fatal(err)
	}
}
