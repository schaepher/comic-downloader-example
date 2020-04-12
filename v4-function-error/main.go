package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
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

func getTitle(html string) (string, error) {
	titleR, err := regexp.Compile(`<title>(.+?)</title>`)
	if err != nil {
		return "", err
	}

	titleMatches := titleR.FindStringSubmatch(html)
	if titleMatches == nil {
		err := errors.New("comic title not found")
		return "", err
	}
	title := titleMatches[1]

	return title, nil
}

func getDownloadId(html string) (int, error) {
	downloadR, err := regexp.Compile(`cn/(\d+)/1.(jpg|png)`)
	if err != nil {
		return 0, err
	}

	downloadMatches := downloadR.FindStringSubmatch(html)
	if downloadMatches == nil {
		err := errors.New("download id not found")
		return 0, err
	}

	downloadId, err := strconv.Atoi(downloadMatches[1])
	if err != nil {
		return 0, err
	}

	return downloadId, nil
}

func main() {
	comicSite := ComicSite{MainPageUrl: "https://*****"}

	res, err := http.Get(comicSite.GetComicMainPageUrl(282526))
	check(err)
	data, err := ioutil.ReadAll(res.Body)
	check(err)
	dataStr := string(data)

	title, err := getTitle(dataStr)
	check(err)
	fmt.Println(title)

	downloadId, err := getDownloadId(dataStr)
	check(err)
	fmt.Println(downloadId)
}
