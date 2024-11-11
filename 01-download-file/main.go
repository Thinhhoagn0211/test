package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	url := flag.String("url", "", "URL to download file from")
	flag.Parse()
	err := DownloadFile(*url)
	if err != nil {
		panic(err)
	}
}

/*
DownloadFile to create a filepath with filename and get file from url and copy into filepath
*/
func DownloadFile(url string) error {
	// Create the file
	out, err := os.Create(strings.Split(url, "/")[len(strings.Split(url, "/"))-1])
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
