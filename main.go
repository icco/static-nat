package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

func main() {
	render_path := "render/"
	base_path := "/Users/nat/Projects/blog-backup/posts/"
	files, err := ioutil.ReadDir(base_path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		log.Printf("%+v", file.Name())
		if strings.HasSuffix(file.Name(), ".md") {
			input, err := ioutil.ReadFile(filepath.Join(base_path, file.Name()))
			if err != nil {
				log.Fatal(err)
			}

			new_filepath := filepath.Join(render_path, "posts", strings.Replace(file.Name(), ".md", "", 1), "index.html")
			err = os.MkdirAll(filepath.Dir(new_filepath), 0777)
			if err != nil {
				log.Fatal(err)
			}

			unsafe := blackfriday.MarkdownCommon(input)
			html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
			err = ioutil.WriteFile(new_filepath, html, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
