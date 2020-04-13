package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type ComicSite struct {
	MainPageUrl string
}

func (cs ComicSite) GetComicMainPageUrl(comicId int) string {
	return fmt.Sprintf("%s/cn/s/%d/", cs.MainPageUrl, comicId)
}

func main() {
	comicSite := ComicSite{
		MainPageUrl: "https://*****",
	}

	// 获取漫画页
	comicMainPageUrl := comicSite.GetComicMainPageUrl(282526)
	res, err := http.Get(comicMainPageUrl)
	check(err)
	data, err := ioutil.ReadAll(res.Body)
	check(err)
	html := string(data)

	// 匹配标题
	titleR, err := regexp.Compile(`<title>(.+?)</title>`)
	check(err)
	titleMatches := titleR.FindStringSubmatch(html)
	if titleMatches == nil {
		panic("comic title not found")
	}
	title := titleMatches[1]
	fmt.Println(title)

	// 匹配下载 ID
	downloadR, err := regexp.Compile(`cn/(\d+)/1.(jpg|png)`)
	check(err)
	downloadMatches := downloadR.FindStringSubmatch(html)
	if downloadMatches == nil {
		panic("download id not found")
	}
	downloadIdStr := downloadMatches[1]
	fmt.Println(downloadIdStr)
}
