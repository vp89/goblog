package main

import (
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func markDowner(args ...interface{}) template.HTML {
	s := blackfriday.MarkdownCommon([]byte(fmt.Sprintf("%s", args...)))
	s = bluemonday.UGCPolicy().SanitizeBytes(s)
	return template.HTML(s)
}

func getMarkdownPreview(res http.ResponseWriter, req *http.Request) {
	md, err := ioutil.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(res, err)
		return
	}

	unsafe := blackfriday.MarkdownCommon(md)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	fmt.Fprint(res, string(html))
}

func titleLinker(args ...interface{}) template.HTML {
	return template.HTML(strings.Replace(fmt.Sprintf("%s", args...), " ", "-", -1))
}

func dateFormatter(args ...interface{}) template.HTML {
	// just trim the datetime string after the seconds
	s := fmt.Sprintf("%s", args...)

	if !strings.Contains(s, ".") {
		return template.HTML("")
	}

	return template.HTML(s[0:strings.Index(s, ".")])
}

func dateFormatterNice(args ...interface{}) template.HTML {
	// just trim the datetime string after the seconds
	s := fmt.Sprintf("%s", args...)

	if !strings.Contains(s, ".") {
		return template.HTML("")
	}

	s = s[0:strings.Index(s, ".")]

	// Go constant Mon Jan 2 15:04:05 MST 2006
	layoutParse := "2006-01-02 15:04:05"
	layoutFormat := "01/02/2006"
	t, err := time.Parse(layoutParse, s)
	checkErr(err)
	s2 := t.Format(layoutFormat)

	return template.HTML(s2)
}
