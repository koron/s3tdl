package main

import (
	"log"
	"os"

	"github.com/koron-go/subcmd"
	"github.com/koron/s3tdl/internal/download"
)

var rootSet = subcmd.DefineRootSet(
	download.Download,
)

func main() {
	if err := subcmd.Run(rootSet, os.Args[1:]...); err != nil {
		log.Fatal(err)
	}
}
