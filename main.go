package main

import (
	"fmt"
	"log"

	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/cli/go-gh"
)

type args struct {
	Url         string `arg:"positional,required"`
	Target      string `arg:"-t" help:"where to put the downloaded diretory"`
	NoRecursive bool   `arg:"--no-recursive" help:"don't download recursively"`
	Depth       int    `arg:"-d" default:"-1" help:"recursive download depth"`
	AsZip       bool   `arg:"--as-zip" help:"download subdirectory as a zip file"`
}

func (args) Version() string {
	return "gh-subdir 0.1.0\nhttps://github.com/brainwo/gh-subdir"
}

func main() {
	var args args
	arg.MustParse(&args)

	depth := args.Depth

	if args.NoRecursive {
		depth = 0
	}

	recursiveDownload(args.Url, depth)

	//TODO: implement logic to bring the download from temporary directory to current directory
	//TODO: implement logic to save directory as zip
}

// Recursively download a directory, once the depth reaches 0 it will not download the subdirectory.
// If the depth is set to -1, it will download all the subfolder.
func recursiveDownload(url string, depth int) {
	schema := strings.SplitN(url, "/", 8)

	if len(schema) < 6 {
		log.Fatal("Invalid url, too short")
	}

	if schema[2] != "github.com" && schema[5] != "tree" {
		log.Fatal("Invalid url")
	}

	owner := schema[3]
	repo := schema[4]
	path := schema[7]

	client, err := gh.RESTClient(nil)
	response := []struct {
		Type         string
		Name         string
		Html_url     string
		Download_url string
	}{}

	err = client.Get(fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path), &response)
	if err != nil {
		log.Fatal(err)
	}

	for _, content := range response {
		switch content.Type {
		case "file":
			downloadFile(content.Download_url)
		case "dir":
			if depth != 0 {
				recursiveDownload(content.Html_url, depth-1)
			}
		// TODO: add supports for submodules and symlinks
		default:
			return
		}
	}
}

// Save file to temp directory
// TODO: implement the downloadFile function
func downloadFile(url string) {
	schema := strings.SplitN(url, "/", 7)

	repo := schema[4]
	path := schema[6]

	// TODO: make it use TMP environment variable
	fmt.Printf("Downloading to: /tmp/gh-subdir/%s/%s\n", repo, path)
	// TODO: make it work for private repositories as well
}
