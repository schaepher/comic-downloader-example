package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

func (cs ComicSite) GetComicIndexUrl(comicId int) string {
	return fmt.Sprintf("%s/cn/%d.js", cs.MainPageUrl, comicId)
}

type ComicFile struct {
	Name string `json:"name"`
}

type Comic struct {
	Id         int
	Title      string
	DownloadId int
	ComicSite  ComicSite
	ComicFiles []ComicFile
}

func (comic *Comic) LoadMeta() error {
	var err error
	var mainPageHtml string

	comicMainPageUrl := comic.ComicSite.GetComicMainPageUrl(comic.Id)
	mainPageHtml, err = comic.getComicMainPageHtml(comicMainPageUrl)
	if err != nil {
		return err
	}

	comic.Title, err = comic.findTitle(mainPageHtml)
	if err != nil {
		return err
	}

	comic.DownloadId, err = comic.findDownloadId(mainPageHtml)
	if err != nil {
		return err
	}

	comicIndexUrl := comic.ComicSite.GetComicIndexUrl(comic.Id)
	comic.ComicFiles, err = comic.readComicIndexes(comicIndexUrl)
	if err != nil {
		return err
	}

	return nil
}

func (_ Comic) getComicMainPageHtml(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (_ Comic) findTitle(html string) (string, error) {
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

func (_ Comic) findDownloadId(html string) (int, error) {
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

func (_ Comic) readComicIndexes(comicIndexUrl string) ([]ComicFile, error) {
	res, err := http.Get(comicIndexUrl)
	if err != nil {
		return nil, err
	}
	htmlByte, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	html := string(htmlByte)

	r, err := regexp.Compile("\\[.+]")
	if err != nil {
		return nil, err
	}
	jsonStr := r.FindString(html)
	validJson := strings.Replace(jsonStr, ",]", "]", 1)

	var pages []ComicFile
	err = json.Unmarshal([]byte(validJson), &pages)
	if err != nil {
		return nil, err
	}

	return pages, nil
}

func main() {
	comicSite := ComicSite{MainPageUrl: "https://*****"}

	comic := &Comic{ComicSite: comicSite, Id: 282526}
	err := comic.LoadMeta()
	check(err)

	fmt.Printf("%+v\n", comic)
}
