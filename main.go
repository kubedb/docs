package main

import (
	"log"

	logs "github.com/appscode/go/log/golog"
	"github.com/kubedb/operator/cmds"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()
	if err := cmds.NewRootCmd("crdReg").Execute(); err != nil {
		log.Fatal(err)
	}
}
