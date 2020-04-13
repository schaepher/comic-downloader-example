package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	url := "https://cn.bing.com"
	res, err := http.Get(url)
	check(err)
	data, err := ioutil.ReadAll(res.Body)
	check(err)

	ioErr := ioutil.WriteFile("cn.bing.com.html", data, 644)
	check(ioErr)

	fmt.Printf("Got:\n%q", string(data))
}
