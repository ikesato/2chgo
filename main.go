package main

import (
	"./nichan"
	"os"
)

func main() {
	arg, err := parseCmdLine()
	if err != nil {
		os.Exit(1)
	}
	nichan.Crawl(arg.Url)
}
