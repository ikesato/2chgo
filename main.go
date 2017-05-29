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
	posts, err := nichan.Crawl(arg.Url)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	if arg.Format == "json" {
		outJSON(posts)
	} else {
		outText(posts)
	}
}

func outJSON(posts []nichan.Post) {
	bytes, err := json.Marshal(posts)
	if err != nil {
		return
	}
	fmt.Println(string(bytes))
}

func outText(posts []nichan.Post) {
	for _, post := range posts {
		fmt.Println("-------------------------------------------")
		fmt.Printf("%v %v: %v (%v)\n", post.Time, post.No, post.Name, post.Uid)
		fmt.Printf("%v\n", post.Message)
	}
}
