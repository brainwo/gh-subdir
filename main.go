package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/cli/go-gh"
	"github.com/cli/go-gh/pkg/api"
)

type args struct {
	Url         string `arg:"positional,required"`
	Target      string `arg:"-t" help:"where to put the downloaded diretory"`
	NoRecursive bool   `arg:"--no-recursive" help:"don't download recursively"`
	Depth       int    `arg:"-d" default:"-1" help:"recursive download depth"`
	// TODO: implement logic to save as zip
	// AsZip       bool   `arg:"--as-zip" help:"download subdirectory as a zip file"`
}

func (args) Version() string {
	return "gh-subdir 0.1.0\nhttps://github.com/brainwo/gh-subdir"
}

type tempFile struct {
	OsFile      *os.File
	Destination string
}

var tempFiles = []tempFile{}

func main() {

	var args args
	arg.MustParse(&args)

	depth := args.Depth

	if args.NoRecursive {
		depth = 0
	}

	recursiveDownload(args.Url, depth)

	// This might be a memory-heavy operation
	// Some alternative methods have been tried with no sucess:
	// - os.Rename, cross device problem
	// - io.Copy, file created but no content
	for _, file := range tempFiles {
		// TODO: create the directory when there is none
		data, err := os.ReadFile(file.OsFile.Name())
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile(file.Destination, data, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		file.OsFile.Close()
	}

	//TODO: implement logic to save directory as zip
}

// Recursively download a directory, once the depth reaches 0 it will not download the subdirectory.
// If the depth is set to -1, it will download all the subdirectory.
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
		case "submodule":
			// TODO: add supports for submodules
		case "symlink":
			// TODO: add supports for symlinks
		default:
			return
		}
	}
}

// Save file to temp directory
func downloadFile(url string) {
	schema := strings.SplitN(url, "/", 7)

	repo := schema[4]
	path := schema[6]

	fmt.Printf("Downloading %s/%s\n", repo, path)
	file, err := os.CreateTemp(os.TempDir(), "gosubdir")
	if err != nil {
		log.Fatal(err)
	}

	opts := api.ClientOptions{}

	client, err := gh.HTTPClient(&opts)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	temp := tempFile{
		OsFile: file,
		// TODO: remove parameter (?=) from path
		Destination: path,
	}

	fmt.Printf("size: %v\n", size)
	tempFiles = append(tempFiles, temp)
}
