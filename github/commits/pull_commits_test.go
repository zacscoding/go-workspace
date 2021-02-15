package commits

import (
	"context"
	"fmt"
	"github.com/google/go-github/v33/github" // with go modules enabled (GO111MODULE=on or outside GOPATH)
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"strings"
	"testing"
)

type SearchOpt struct {
	User       string
	Repo       string
	Keywords   []string
	MaxCommits int
}

func Test1(t *testing.T) {
	opt := SearchOpt{
		User:       "zacscoding",
		Repo:       "go-workspace",
		Keywords:   []string{"test", "gorm"},
		MaxCommits: 5,
	}
	writer := table.NewWriter()
	writer.SetOutputMirror(os.Stdout)
	err := searchGithubCommitMessages(&opt, writer)
	assert.NoError(t, err)
}

func searchGithubCommitMessages(opt *SearchOpt, writer table.Writer) error {
	var (
		client      = github.NewClient(nil)
		page        = 0
		perPage     = 100
		remain      = opt.MaxCommits
		readCommits = 0
		idx         = 1
		continueReq = true
	)
	if opt.MaxCommits > 0 {
		perPage = opt.MaxCommits
	}

	writer.Style().Format.Header = text.FormatDefault
	writer.Style().Format.Footer = text.FormatDefault
	writer.SetTitle("Matched Commits from /%s/%s. Keywords:%s", opt.User, opt.Repo, strings.Join(opt.Keywords, ","))
	writer.AppendHeader(table.Row{"#", "Message", "URL"})

	for {
		if !continueReq {
			break
		}
		commits, response, err := client.Repositories.ListCommits(context.Background(),
			opt.User,
			opt.Repo,
			&github.CommitsListOptions{
				ListOptions: github.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			})
		if err != nil {
			return err
		}
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch commits. status code:%d", response.StatusCode)
		}
		for _, commit := range commits {
			var (
				include = false
				msg     = strings.ToLower(commit.GetCommit().GetMessage())
			)
			readCommits++

			for _, word := range opt.Keywords {
				if strings.Contains(msg, word) {
					include = true
					break
				}
			}

			if include {
				writer.AppendRow(table.Row{
					idx, commit.GetCommit().GetMessage(), fmt.Sprintf("https://github.com/%s/%s/commit/%s", opt.User, opt.Repo, commit.GetSHA()),
				})
				writer.AppendSeparator()
				idx++
				remain--
			}
			if opt.MaxCommits > 0 && remain < 0 {
				continueReq = false
				break
			}
		}
		if response.NextPage == 0 {
			continueReq = false
		}
		page = response.NextPage
	}
	writer.AppendFooter(table.Row{"", "", fmt.Sprintf("Total Read Commits: #%d", readCommits)})
	writer.Render()
	return nil
}
