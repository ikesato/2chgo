package nichan

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Post struct {
	No      int
	Name    string
	Message string
	Uid     string
	Time    time.Time
}

func Crawl(url string) ([]Post, error) {
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer res.Body.Close()

	utfBody := transform.NewReader(bufio.NewReader(res.Body), japanese.ShiftJIS.NewDecoder())
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	posts := []Post{}
	doc.Find("div.post").Each(func(i int, s *goquery.Selection) {
		var errno, errtime error
		post := Post{}
		post.No, errno = strconv.Atoi(strings.TrimSpace(s.Find("div.meta span.number").Text()))
		post.Name = strings.TrimSpace(s.Find("div.meta span.name").Text())
		post.Time, errtime = time.Parse("2017/05/27(土) 14:23:15.40 997", strings.TrimSpace(s.Find("div.meta span.date").Text()))
		post.Uid = strings.TrimSpace(s.Find("div.meta span.uid").Text())
		if errno != nil || errtime != nil {
			fmt.Println("parse error")
			fmt.Println(errno)
			fmt.Println(errtime)
		}

		m := s.Find("div.message span")
		m.Find("a").Each(func(_ int, link *goquery.Selection) {
			text := strings.TrimSpace(link.Text())
			link.ReplaceWithHtml(text)
		})

		text, _ := m.Html()
		re, _ := regexp.Compile("\\s*\\<br/\\>\\s*")
		text = re.ReplaceAllString(text, "\n")
		text = strings.TrimSpace(text)
		re, _ = regexp.Compile("\\<[\\S\\s]+?\\>") // remove all tags
		text = re.ReplaceAllString(text, "")
		text = strings.TrimSpace(text)
		//fmt.Println("-------------------------------------------")
		//fmt.Printf("%v %v: %v (%v)\n", date, no, name, uid)
		//fmt.Printf("%v\n", text)
		post.Message = text
	})
	return posts, nil
}
