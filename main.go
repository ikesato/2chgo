package main

import (
	"./nichan"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	arg, err := parseCmdLine()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	thread, err := nichan.Crawl(arg.Url)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	if arg.Format == "json" {
		outJSON(thread)
	} else {
		outText(thread)
	}
}

func outJSON(thread *nichan.Thread) {
	bytes, err := json.Marshal(thread)
	if err != nil {
		return
	}
	fmt.Println(string(bytes))
}

func outText(thread *nichan.Thread) {
	fmt.Println("Title     : ", thread.Title)
	fmt.Println("Next URL  : ", thread.NextURL)
	fmt.Println("Board URL : ", thread.BoardURL)
	for _, post := range thread.Posts {
		fmt.Println("-------------------------------------------")
		fmt.Printf("%v %v: %v (%v)\n", post.Time, post.No, post.Name, post.Uid)
		fmt.Printf("%v\n", post.Message)
	}
}
