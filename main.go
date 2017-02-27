package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"github.com/spf13/cast"
	"github.com/spf13/hugo/parser"
)

type Post struct {
	Title    string
	Html     []byte
	Md       []byte
	DateTime time.Time
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Not enough arguments. Need one argument with a path to files.")
	}

	render_path := "render/"
	base_path := os.Args[1]
	files, err := ioutil.ReadDir(base_path)
	if err != nil {
		log.Fatal(err)
	}

	// Write out posts
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
			post, err := parseMarkdown(input, file)
			if err != nil {
				log.Fatal(err)
			}
			err = ioutil.WriteFile(new_filepath, post.Html, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func parseMarkdown(input []byte, file os.FileInfo) (*Post, error) {
	inc := twitterHandleToMarkdown(input)
	inc = hashTagsToMarkdown(inc)

	// get the page from file
	p, err := parser.ReadFrom(bytes.NewReader(input))
	if err != nil {
		log.Printf("Error parsing file %v: %v", file.Name(), err.Error())
		return nil, err
	}

	meta_uncast, err := p.Metadata()
	if err != nil {
		log.Printf("Error getting metadata from %v: %v", file.Name(), err.Error())
		return nil, err
	}

	meta := map[string]string{}
	if meta_uncast != nil {
		meta, err = cast.ToStringMapStringE(meta_uncast)
		if err != nil {
			log.Printf("Error casting metadata for %v: %v. Metadata: %+v", file.Name(), err.Error(), meta_uncast)
			return nil, err
		}
	}

	unsafe := blackfriday.MarkdownCommon(p.Content())
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	datetime, err := time.Parse("2006-01-02 15:04:05 -0700 MST", meta["datetime"])
	if err != nil {
		return nil, err
	}

	return &Post{
		Title:    meta["title"],
		Md:       p.Content(),
		Html:     html,
		DateTime: datetime,
	}, nil
}

var TwitterHandleRegex *regexp.Regexp = regexp.MustCompile(`(\s)@([_A-Za-z0-9]+)`)

func twitterHandleToMarkdown(in []byte) []byte {
	return TwitterHandleRegex.ReplaceAll(in, []byte("$1[@$2](http://twitter.com/$2)"))
}

var HashtagRegex *regexp.Regexp = regexp.MustCompile(`(\s)#(\w+)`)

func hashTagsToMarkdown(in []byte) []byte {
	return HashtagRegex.ReplaceAll(in, []byte("$1[#$2](/tags/$2)"))
}
