package main

import (
	"./thread"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	MainPageUrl        string
	DownloadSourceUrls []string
}

func (cs ComicSite) GetComicMainPageUrl(comicId int) string {
	return fmt.Sprintf("%s/cn/s/%d/", cs.MainPageUrl, comicId)
}

func (cs ComicSite) GetComicIndexUrl(comicId int) string {
	return fmt.Sprintf("%s/cn/%d.js", cs.MainPageUrl, comicId)
}

func (cs ComicSite) GetComicDownloadUrl(comicId int, filename string, srcIndex int) (string, error) {
	if srcIndex < 0 || srcIndex >= len(cs.DownloadSourceUrls) {
		return "", errors.New("no more urls")
	}

	return fmt.Sprintf("%s/%d/%s", cs.DownloadSourceUrls[srcIndex], comicId, filename), nil
}

type ComicFile struct {
	Name string `json:"name"`
}

type Comic struct {
	Id            int
	Title         string
	DownloadId    int
	ComicSite     ComicSite
	ComicFiles    []ComicFile
	ComicsRootDir string
}

func (comic Comic) GetDirPath() string {
	return comic.ComicsRootDir + "/" + strconv.Itoa(comic.Id)
}

func (comic Comic) GetFilePath(filename string) string {
	return comic.GetDirPath() + "/" + filename
}

func (comic Comic) GetMetaFilePath() string {
	return comic.GetDirPath() + "/meta.json"
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
		err := errors.New("downloadComic id not found")
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

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadImg(comic Comic, comicFile ComicFile) {
	log.Printf("Start downloading: %s\n", comicFile.Name)
	for i := 0; i < len(comic.ComicSite.DownloadSourceUrls); i++ {
		downloadUrl, err := comic.ComicSite.GetComicDownloadUrl(comic.DownloadId, comicFile.Name, i)
		if err != nil {
			break
		}

		log.Printf("Trying: %s\n", downloadUrl)
		resp, err := http.Get(downloadUrl)
		if err != nil || resp.StatusCode != 200 {
			log.Printf("Failed: %s\n", downloadUrl)
			continue
		}
		data, err := ioutil.ReadAll(resp.Body)

		err = ioutil.WriteFile(comic.GetFilePath(comicFile.Name), data, 0644)
		if err != nil {
			log.Printf("Failed : %s\n", comicFile.Name)
			return
		}

		log.Printf("Saved : %s\n", comic.GetFilePath(comicFile.Name))
		return
	}
}

type DownloadParam struct {
	Comic     Comic
	ComicFile ComicFile
}

func downloadComic(comic Comic, maxThread int) error {
	log.Printf("Downloading: %s\n", comic.Title)

	err := createDirIfNotExist(comic.GetDirPath())
	if err != nil {
		return err
	}

	data, err := json.Marshal(comic)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(comic.GetMetaFilePath(), data, 0644)
	if err != nil {
		return err
	}
	log.Printf("Meta file saved: %s\n", comic.GetMetaFilePath())

	existFiles, err := ListDirFiles(comic.GetDirPath())
	if err != nil {
		return err
	}

	log.Println("Downloading comic files")
	tp := Thread.Pool{MaxThread: maxThread}
	tp.Prepare(func(param interface{}) {
		downloadParam := param.(DownloadParam)
		downloadImg(downloadParam.Comic, downloadParam.ComicFile)
	})

	for _, comicFile := range comic.ComicFiles {
		if InArray(comicFile.Name, existFiles) {
			continue
		}
		tp.RunWith(DownloadParam{Comic: comic, ComicFile: comicFile})
	}

	tp.Wait()

	return nil
}

func ListDirFiles(root string) ([]string, error) {
	var files []string
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}
	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}

func InArray(item string, items []string) bool {
	for _, tmpItem := range items {
		if tmpItem == item {
			return true
		}
	}
	return false
}

type Config struct {
	MainPageUrl        string   `json:"mainPageUrl"`
	DownloadSourceUrls []string `json:"downloadSourceUrls"`
	MaxThread          int      `json:"maxThread"`
	ComicIds           []int    `json:"comicIds"`
	ComicsRootDir      string   `json:"comicsRootDir"`
}

func (config *Config) Load(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, config)
	return err
}

func main() {
	var err error

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)

	config := Config{}
	err = config.Load(dir + "/config.json")
	check(err)
	if config.ComicsRootDir == "" {
		config.ComicsRootDir = dir
	}

	comicSite := ComicSite{
		MainPageUrl:        config.MainPageUrl,
		DownloadSourceUrls: config.DownloadSourceUrls,
	}

	for _, comicId := range config.ComicIds {
		comic := &Comic{ComicSite: comicSite, Id: comicId, ComicsRootDir: config.ComicsRootDir}
		err = comic.LoadMeta()
		check(err)

		err = downloadComic(*comic, config.MaxThread)
		check(err)

		log.Printf("Downloaded: %d", comic.Id)
	}
}
