package main

import (
	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/crawler"
)

func main() {
	cmd.Parse()
	crawler.Run()
}
