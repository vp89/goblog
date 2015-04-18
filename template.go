package main

import (
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"html/template"
	"strings"
	"time"
)

func markDowner(args ...interface{}) template.HTML {
	s := blackfriday.MarkdownCommon([]byte(fmt.Sprintf("%s", args...)))
	s = bluemonday.UGCPolicy().SanitizeBytes(s)
	return template.HTML(s)
}

func titleLinker(args ...interface{}) template.HTML {
	return template.HTML(strings.Replace(fmt.Sprintf("%s", args...), " ", "-", -1))
}

func dateFormatter(args ...interface{}) template.HTML {

	t, _ := time.Parse("20060102", fmt.Sprintf("%s", args...))
	fmt.Println(t.String())
	return template.HTML(t.String())
}
