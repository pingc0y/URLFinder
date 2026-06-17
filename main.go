package main

import (
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/crawler"
	"github.com/pingc0y/URLFinder/util"
	"io"
	"log"
	"os"
)

func main() {
	log.SetOutput(io.Discard)
	if !cmd.HelpRequested(os.Args[1:]) {
		util.GetUpdate()
	}
	cmd.Parse()
	crawler.Run()
}
