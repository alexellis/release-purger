package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v36/github"
	"golang.org/x/oauth2"
)

func main() {
	var (
		owner          string
		repo           string
		tokenFile      string
		purgeArtifacts bool
		purgeReleases  bool
	)

	flag.StringVar(&owner, "owner", "inlets", "username or org on GitHub")
	flag.StringVar(&repo, "repo", "inlets-archived", "repo on GitHub")
	flag.StringVar(&tokenFile, "token-file", "./token", "path to personal access token saved as a file")
	flag.BoolVar(&purgeArtifacts, "purge-artifacts", false, "purge all release artifacts")
	flag.BoolVar(&purgeReleases, "purge-releases", false, "purge the release itself")

	flag.Parse()

	t, err := os.ReadFile(tokenFile)
	if err != nil {
		panic(err)
	}

	token := strings.TrimSpace(string(t))

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	fmt.Printf("Looking up %s/%s\n", owner, repo)

	listOpts := &github.ListOptions{PerPage: 100, Page: 1}

	releases, rinfo, err := client.Repositories.ListReleases(context.Background(), owner, repo, listOpts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d releases\n", len(releases))

	log.Printf("Next page: %s", rinfo.NextPageToken)

	for _, release := range releases {
		fmt.Printf("Releases: %s (%d)\n", release.GetName(), release.GetID())

		assets, _, _ := client.Repositories.ListReleaseAssets(context.Background(),
			owner, repo, release.GetID(), &github.ListOptions{
				Page:    0,
				PerPage: 100,
			})

		for _, asset := range assets {
			fmt.Println(asset.ID, asset.GetName(),
				asset.BrowserDownloadURL, asset.GetContentType(),
				asset.GetDownloadCount(), asset.Label)

			if purgeArtifacts {
				_, err := client.Repositories.DeleteReleaseAsset(context.Background(), owner, repo, asset.GetID())
				if err != nil {
					panic(err)
				}
				fmt.Printf("Artifact: %d\t%s deleted\n", asset.GetID(), asset.GetName())
			}
		}

		if purgeReleases {
			_, err := client.Repositories.DeleteRelease(context.Background(), owner, repo, release.GetID())
			if err != nil {
				panic(err)
			}
			fmt.Printf("Release: %d\t%s deleted\n", release.GetID(), release.GetName())
		}
	}
}
