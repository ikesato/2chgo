package nichan

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DEBUG = false

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
	titleBase := resolveTitleBase(thread.Title)

	doc.Find(".menubottommenu a.menuitem").EachWithBreak(func(i int, s *goquery.Selection) bool {
		match, _ := regexp.MatchString("掲示板に戻る", s.Text())
		if match {
			thread.BoardURL = strings.TrimSpace(s.AttrOr("href", ""))
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

		nextURL := findNextURL(m, titleBase, post.No)
		if len(nextURL) > 0 {
			thread.NextURL = nextURL
		}

		// <a>
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
	thread.BoardURL = normalizeUrlScheme(thread.BoardURL, url)
	thread.NextURL = normalizeUrlScheme(thread.NextURL, url)
	return thread, nil
}

func resolveTitleBase(title string) string {
	title = regexp.MustCompile("\\d+[\\D]*$").ReplaceAllString(title, "")
	title = regexp.MustCompile("[０-９]+[^０-９]*$").ReplaceAllString(title, "")
	title = regexp.MustCompile("[一二三四五六七八九〇十百千万零壱弐参肆伍陸漆捌玖拾壹弌貳貮參弎質柒百陌佰]+[^一二三四五六七八九〇十百千万零壱弐参肆伍陸漆捌玖拾壹弌貳貮參弎質柒百陌佰]*$").ReplaceAllString(title, "")
	return title
}

func findNextURL(m *goquery.Selection, titleBase string, no int) string {
	if no <= 900 {
		return ""
	}
	text := m.Text()
	if strings.Index(text, "次") == -1 {
		return ""
	}
	if strings.Index(text, titleBase) == -1 {
		return ""
	}
	return m.Find("a").Last().AttrOr("href", "")
}

func normalizeUrlScheme(targetURL, baseURL string) string {
	burl, _ := url.Parse(baseURL)
	if strings.HasPrefix(targetURL, "//") {
		return burl.Scheme + ":" + targetURL
	}
	turl, _ := url.Parse(targetURL)
	if turl.Scheme == "http" && burl.Scheme == "https" {
		return regexp.MustCompile("^http:").ReplaceAllString(targetURL, "https:")
	}
	return targetURL
}
