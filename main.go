package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
)

func main() {
	var (
		owner          string
		repo           string
		tokenFile      string
		purgeArtifacts bool
		purgeReleases  bool
		dryRun         bool
		maxDays        int
	)

	flag.StringVar(&owner, "owner", "", "username or org on GitHub")
	flag.StringVar(&repo, "repo", "", "repo on GitHub")
	flag.StringVar(&tokenFile, "token-file", "./token", "path to personal access token saved as a file")
	flag.BoolVar(&purgeArtifacts, "purge-artifacts", false, "purge all release artifacts")
	flag.BoolVar(&purgeReleases, "purge-releases", false, "purge the release itself")
	flag.BoolVar(&dryRun, "dry-run", true, "dry-run")
	flag.IntVar(&maxDays, "max-days", 120, "maximum amount of days in age of kept releases")

	flag.Parse()

	if len(owner) == 0 {
		panic("owner is required")
	}
	if len(repo) == 0 {
		panic("repo is required")
	}

	maxAge := time.Duration(maxDays) * 24 * time.Hour

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

	releases, rinfo, err := client.Repositories.ListReleases(context.Background(),
		owner,
		repo,
		listOpts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d releases\n", len(releases))

	log.Printf("Next page: %s", rinfo.NextPageToken)

	maxDate := time.Now().Add(-maxAge)
	for _, release := range releases {
		fmt.Printf(" - %s:\t%s (%d)\n", release.GetTagName(), release.GetName(), release.GetID())

		if purgeArtifacts {
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
		}

		if purgeReleases {

			fmt.Printf("Max date: %v, release date: %v\n", maxDate.Round(time.Hour), release.GetCreatedAt().Time.Round(time.Hour))
			if release.CreatedAt.Time.Before(maxDate) {
				fmt.Printf("Deleting release: %s, name: %s, created: %s, age: %s\n",
					release.GetTagName(), release.GetName(), release.GetCreatedAt(),
					time.Since(release.GetCreatedAt().Time).Round(time.Hour))

				fmt.Printf("Deleting: %v now\n", release.GetTagName())
				time.Sleep(time.Millisecond * 100)

				if !dryRun {
					if _, err := client.Repositories.DeleteRelease(context.Background(), owner, repo, release.GetID()); err != nil {
						panic(fmt.Sprintf("unable to delete tag: %s, error: %s", release.GetTagName(), err.Error()))

					}

					tagRef := fmt.Sprintf("refs/tags/%s", release.GetTagName())
					if _, err := client.Git.DeleteRef(context.Background(), owner, repo, tagRef); err != nil {
						panic(fmt.Sprintf("unable to delete tag reference: %s, error: %s", tagRef, err.Error()))
					}
				}
				log.Printf("%s", release.GetTagName())

				fmt.Printf("Release: %d\t%s deleted\n", release.GetID(), release.GetName())
			} else {
				fmt.Printf("Skipping release: %s, created: %s, age: %s\n", release.GetName(),
					release.GetCreatedAt(),
					time.Since(release.GetCreatedAt().Time).Round(time.Hour))
			}
		}
	}
}
