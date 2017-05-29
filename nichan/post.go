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
		if post.No > 1000 {
			return
		}
		post.Name = strings.TrimSpace(s.Find("div.meta span.name").Text())
		t := strings.TrimSpace(s.Find("div.meta span.date").Text())
		re, _ := regexp.Compile("\\(.?\\)")
		t = re.ReplaceAllString(t, "")
		post.Time, errtime = time.Parse("2006/01/02 15:04:05.00", t)
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

		// "<br/>" -> "\n"
		re = regexp.MustCompile("\\s*\\<br/\\>\\s*")
		text = re.ReplaceAllString(text, "\n")

		// remove all tags
		re = regexp.MustCompile("\\<[\\S\\s]+?\\>")
		text = re.ReplaceAllString(text, "")

		// "&gt;" -> ">"
		re = regexp.MustCompile("&gt;")
		text = re.ReplaceAllString(text, ">")

		// "&lt;" -> "<"
		re = regexp.MustCompile("&lt;")
		text = re.ReplaceAllString(text, "<")

		// "&amp;" -> "&"
		re = regexp.MustCompile("&amp;")
		text = re.ReplaceAllString(text, "&")

		// "&#xx;"
		re = regexp.MustCompile("&#(\\d+);")
		for {
			group := re.FindStringSubmatch(text)
			if group == nil {
				break
			}
			num, _ := strconv.Atoi(group[1])
			text = regexp.MustCompile("&#\\d+;").ReplaceAllString(text, string(num))
		}

		text = strings.TrimSpace(text)
		post.Message = text
		posts = append(posts, post)
	})
	return posts, nil
}
