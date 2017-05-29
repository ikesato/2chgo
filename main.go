package main

import (
	"./nichan"
	"fmt"
	"os"
)

func main() {
	arg, err := parseCmdLine()
	if err != nil {
		os.Exit(1)
	}
	posts, err := nichan.Crawl(arg.Url)
	fmt.Println(len(posts))
	fmt.Println(err)
	if err == nil {
		for _, post := range posts {
			fmt.Println("-------------------------------------------")
			fmt.Printf("%v %v: %v (%v)\n", post.Time, post.No, post.Name, post.Uid)
			fmt.Printf("%v\n", post.Message)
		}
	}
}
