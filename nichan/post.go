package nichan

import (
	"bufio"
	"errors"
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

const DEBUG = true

type Post struct {
	No      int       `json:"no"`
	Name    string    `json:"name"`
	Message string    `json:"message"`
	Uid     string    `json:"uid"`
	Time    time.Time `json:"time"`
}

type Thread struct {
	Title    string
	BoardURL string
	NextURL  string
	Posts    []Post
}

func Crawl(url string) (*Thread, error) {
	var utfBody *transform.Reader
	var err error
	if !DEBUG {
		res, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			return nil, errors.New("Error: failed to http, URL => " + url)
		}

		defer res.Body.Close()

		utfBody = transform.NewReader(bufio.NewReader(res.Body), japanese.ShiftJIS.NewDecoder())
		if err != nil {
			fmt.Println(err)
			return nil, errors.New("Error: failed to convert to SJIS, URL => " + url)
		}
	} else {
		fp, err := os.OpenFile("b.html", os.O_RDONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return nil, errors.New("Error: failed to open file")
		}
		utfBody = transform.NewReader(bufio.NewReader(fp), japanese.ShiftJIS.NewDecoder())
	}

	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("Error: failed to parse HTML, URL => " + url)
	}

	var thread *Thread = &Thread{}
	thread.Title = strings.TrimSpace(doc.Find("h1").Text())
	thread.Title = regexp.MustCompile("\\[無断転載禁止\\]©2ch.net$").ReplaceAllString(thread.Title, "")
	thread.Title = strings.TrimSpace(thread.Title)

	doc.Find(".menubottommenu a.menuitem").EachWithBreak(func(i int, s *goquery.Selection) bool {
		match, _ := regexp.MatchString("掲示板に戻る", s.Text())
		if match {
			thread.BoardURL = strings.TrimSpace(s.AttrOr("href", ""))
			if strings.HasPrefix(thread.BoardURL, "//") {
				thread.BoardURL = "https:" + thread.BoardURL
			}
			return false
		}
		return true
	})

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
		post.Uid = regexp.MustCompile("^ID:").ReplaceAllString(post.Uid, "")
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
		thread.Posts = append(thread.Posts, post)
	})
	return thread, nil
}
