package main

import (
	"errors"
	"fmt"
	"github.com/droundy/goopt"
)

type Args struct {
	Url string
}

func parseCmdLine() (Args, error) {
	goopt.Description = func() string {
		return "Crawl 2ch thread and output various format."
	}
	goopt.Version = "1.0"
	goopt.Summary = "2chgo [Options..] URL"
	var format = goopt.Alternatives([]string{"-f", "--color"},
		[]string{"text", "json"},
		"Specify the format of output, default format is text")
	goopt.Parse(nil)

	var arg Args
	if len(goopt.Args) == 0 {
		fmt.Println(goopt.Usage())
		return arg, errors.New("need URL")
	}
	fmt.Println("aaaaaaaaaaaa", *format)

	arg.Url = goopt.Args[0]
	return arg, nil
}
