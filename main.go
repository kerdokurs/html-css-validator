package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var client *http.Client

const (
	HTML_VAL_URL = "https://validator.w3.org/nu/?out=json"
	CSS_VAL_URL  = "https://jigsaw.w3.org/css-validator/validator"
)

type HTTPValResult struct {
	Messages []string `json:"messages"`
}

// https://stackoverflow.com/a/55300382/6660721
func walkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func validateCSS(rootPath string) {
	cssFiles, err := walkMatch(rootPath, "*.css")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Viga CSS failide laadimisel\n")
		panic(err)
	}

	for _, file := range cssFiles {
		contentBytes, err := os.ReadFile(file)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Viga faili %s lugemisel\n", file)
			continue
		}

		content := string(contentBytes)

		req, err := http.NewRequest("GET", CSS_VAL_URL+"?text="+url.QueryEscape(content), nil)

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Viga päringu loomisel faili %s jaoks\n", file)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Viga faili %s valideerimisel\n", file)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		if strings.Contains(string(body), "Congratulations! No Error Found.") {
			fmt.Printf("[%s]\t  OK\n", file)
		} else {
			fmt.Printf("[%s]\tFAIL\n", file)
		}

		time.Sleep(time.Second)
	}
}

func validateHTML(rootPath string) {
	htmlFiles, err := walkMatch(rootPath, "*.html")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Viga HTML failide laadimisel\n")
		panic(err)
	}

	for _, file := range htmlFiles {
		contentBytes, err := os.ReadFile(file)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Viga faili %s lugemisel\n", file)
			continue
		}

		content := string(contentBytes)

		form := url.Values{}
		form.Add("content", content)

		req, err := http.NewRequest("POST", HTML_VAL_URL, bytes.NewBuffer(contentBytes))
		req.Header.Add("Content-Type", "text/html")

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Viga päringu loomisel faili %s jaoks\n", file)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Viga faili %s valideerimisel\n", file)
			continue
		}

		target := new(HTTPValResult)

		json.NewDecoder(resp.Body).Decode(target)

		if len(target.Messages) == 0 {
			fmt.Printf("[%s]\t  OK\n", file)
		} else {
			fmt.Printf("[%s]\tFAIL\n", file)
		}

		time.Sleep(time.Second)
	}
}

func main() {
	client = &http.Client{}

	if len(os.Args) < 2 {
		fmt.Printf("Kasutusviis: %s <kaust>\n", os.Args[0])
		return
	}

	rootPath := os.Args[1]

	validateCSS(rootPath)
	validateHTML(rootPath)
}
