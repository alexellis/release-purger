package main

import (
	"context"
	"fmt"

	"github.com/google/go-github/v36/github"
)

func main() {
	client := github.NewClient(nil)

	owner := "inlets"
	repo := "inlets"
	fmt.Printf("Looking up %s/%s\n", owner, repo)

	listOpts := &github.ListOptions{PerPage: 1, Page: 1}

	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, listOpts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d releases\n", len(releases))

	for _, release := range releases {
		fmt.Printf("Releases: %s (%d)\n", release.GetName(), release.GetID())

		assets, _, _ := client.Repositories.ListReleaseAssets(context.Background(),
			owner, repo, release.GetID(), &github.ListOptions{
				Page:    0,
				PerPage: 100,
			})

		for _, a := range assets {
			fmt.Println(a.ID, a.GetName(), a.BrowserDownloadURL, a.GetContentType(), a.GetDownloadCount(), a.Label)
		}
	}
}
