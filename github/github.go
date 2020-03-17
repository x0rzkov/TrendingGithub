package github

import (
	"context"
	"os"
	"net/http"

	"golang.org/x/oauth2"
	"github.com/k0kubun/pp"
	"github.com/google/go-github/v29/github"
)

// Repository represents a single repository from github.
// This struct is a stripped down version of github.Repository.
// We only return the values we need here.
type Repository struct {
	StargazersCount *int `json:"stargazers_count,omitempty"`
	Topics []string `json:"topics,omitempty"`
}

// GetRepositoryDetails will retrieve details about the repository owner/repo from github.
func GetRepositoryDetails(owner, repo string) (*Repository, error) {

	var tc *http.Client
	tc = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	))

	client := github.NewClient(tc)

	repository, _, err := client.Repositories.Get(context.Background(), owner, repo)
	if repository == nil {
		return nil, err
	}


	pp.Println(repository.Topics)

	r := &Repository{
		StargazersCount: repository.StargazersCount,
		Topics: repository.Topics,
	}
	return r, err
}
