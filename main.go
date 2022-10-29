package main

import (
	"fmt"
	"log"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
)

func main() {
	var args struct {
		Repo string `arg:"positional"`
	}
	arg.MustParse(&args)
	repo := strings.Split(args.Repo, "/")
	if len(repo) != 2 {
		fmt.Println("Error parsing repository name, use owner/repository format")
		return
	}

	client, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}
	var query struct {
		Repository struct {
			Refs struct {
				Nodes []struct {
					Name string
				}
			} `graphql:"refs(refPrefix: $refPrefix, first: $first)"`
			StargazerCount int
		} `graphql:"repository(owner: $owner, name: $name)"`
	}
	variables := map[string]interface{}{
		"first":     graphql.Int(100),
		"refPrefix": graphql.String("refs/heads/"),
		"owner":     graphql.String(repo[0]),
		"name":      graphql.String(repo[1]),
	}
	err = client.Query("RepositoryIssues", &query, variables)
	if err != nil {
		fmt.Println("Repository not found")
		log.Fatal(err)
	}
	fmt.Println(query.Repository.Refs.Nodes)
}

// https://github.com/cli/go-gh/blob/trunk/example_gh_test.go
