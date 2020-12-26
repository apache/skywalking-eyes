//
// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//
package review

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	comments2 "github.com/apache/skywalking-eyes/license-eye/pkg/comments"
	config2 "github.com/apache/skywalking-eyes/license-eye/pkg/config"
	header2 "github.com/apache/skywalking-eyes/license-eye/pkg/header"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

var (
	Identification = "license-eye hidden identification"

	gh  *github.Client
	ctx context.Context

	owner string
	repo  string
	sha   string
	pr    int

	requiredEnvVars = []string{
		"GITHUB_TOKEN",
		"GITHUB_HEAD_REF",
		"GITHUB_REPOSITORY",
		"GITHUB_EVENT_NAME",
		"GITHUB_EVENT_PATH",
	}
)

func init() {
	if os.Getenv("INPUT_GITHUB_TOKEN") == "" {
		logger.Log.Warnln("GITHUB_TOKEN is not set, license-eye won't comment on the pull request")
		return
	}

	if !IsPR() {
		return
	}
	if !IsGHA() {
		panic(fmt.Errorf(fmt.Sprintf(
			`this must be run on GitHub Actions or you have to set the environment variables %v manually.`, requiredEnvVars,
		)))
	}

	s, err := GetSha()
	if err != nil {
		logger.Log.Warnln("failed to get sha", err)
		return
	}

	sha = s
	token := os.Getenv("GITHUB_TOKEN")
	ref := os.Getenv("GITHUB_REF")
	fullName := os.Getenv("GITHUB_REPOSITORY")
	logger.Log.Debugln("ref:", ref, "; repo:", fullName, "; sha:", sha)
	ownerRepo := strings.Split(fullName, "/")
	if len(ownerRepo) != 2 {
		logger.Log.Warnln("Length of ownerRepo is not 2", ownerRepo)
		return
	}
	owner, repo = ownerRepo[0], ownerRepo[1]
	matches := regexp.MustCompile(`refs/pull/(\d+)/merge`).FindStringSubmatch(ref)
	if len(matches) < 1 {
		logger.Log.Warnln("Length of ref < 1", matches)
		return
	}
	prString := matches[1]
	if p, err := strconv.Atoi(prString); err == nil {
		pr = p
	} else {
		logger.Log.Warnln("Failed to parse pull request number.", err)
		return
	}

	ctx = context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	gh = github.NewClient(tc)
}

// Header reviews the license header, including suggestions on the pull request and an overview of the checks.
func Header(result *header2.Result, config *config2.Config) error {
	if !result.HasFailure() || !IsPR() || gh == nil || config.Header.Comment == header2.Never {
		return nil
	}

	commentedFiles := make(map[string]bool)
	for _, comment := range GetAllReviewsComments() {
		decodeString := comment.GetBody()
		if strings.Contains(decodeString, Identification) {
			logger.Log.Debugln("Path:", comment.GetPath())
			commentedFiles[comment.GetPath()] = true
		}
	}
	logger.Log.Debugln("CommentedFiles:", commentedFiles)

	s := "RIGHT"
	l := 1

	var comments []*github.DraftReviewComment
	for _, changedFile := range GetChangedFiles() {
		logger.Log.Debugln("ChangedFile:", changedFile.GetFilename())
		if commentedFiles[changedFile.GetFilename()] {
			logger.Log.Debugln("ChangedFile was reviewed, skipping:", changedFile.GetFilename())
			continue
		}
		for _, invalidFile := range result.Failure {
			if !strings.HasSuffix(invalidFile, changedFile.GetFilename()) {
				continue
			}
			blob, _, err := gh.Git.GetBlob(ctx, owner, repo, changedFile.GetSHA())
			if err != nil {
				logger.Log.Warnln("Failed to get blob:", changedFile.GetFilename(), changedFile.GetSHA())
				continue
			}
			header, err := header2.GenerateLicenseHeader(comments2.FileCommentStyle(changedFile.GetFilename()), &config.Header)
			if err != nil {
				logger.Log.Warnln("Failed to generate comment header:", changedFile.GetFilename())
				continue
			}
			decodeString, err := base64.StdEncoding.DecodeString(blob.GetContent())
			if err != nil {
				logger.Log.Debugln("Failed to decode blob content:", err)
				continue
			}
			body := "```suggestion\n" + header + strings.Split(string(decodeString), "\n")[0] + "\n```\n" + fmt.Sprintf(`<!-- %v -->`, Identification)
			comments = append(comments, &github.DraftReviewComment{
				Path: changedFile.Filename,
				Body: &body,
				Side: &s,
				Line: &l,
			})
		}
	}

	if err := tryReview(result, config, comments); err != nil {
		return err
	}

	return nil
}

func tryReview(result *header2.Result, config *config2.Config, comments []*github.DraftReviewComment) error {
	tryBestEffortToComment := func() error {
		if err := doReview(result, comments); err != nil {
			logger.Log.Warnln("Failed to create review comment, fallback to a plain comment:", err)
			_ = doReview(result, nil)
			return err
		}
		return nil
	}

	if config.Header.Comment == header2.Always {
		if err := tryBestEffortToComment(); err != nil {
			return err
		}
	} else if config.Header.Comment == header2.OnFailure && len(comments) > 0 {
		if err := tryBestEffortToComment(); err != nil {
			return err
		}
	}
	return nil
}

func doReview(result *header2.Result, comments []*github.DraftReviewComment) error {
	logger.Log.Debugln("Comments:", comments)

	c := Markdown(result)
	e := "COMMENT"
	if _, _, err := gh.PullRequests.CreateReview(ctx, owner, repo, pr, &github.PullRequestReviewRequest{
		CommitID: &sha,
		Body:     &c,
		Event:    &e,
		Comments: comments,
	}); err != nil {
		return err
	}
	return nil
}

// GetChangedFiles returns the changed files in this pull request.
func GetChangedFiles() []*github.CommitFile {
	prsvc := gh.PullRequests
	options := &github.ListOptions{Page: 1, PerPage: 100}

	var allFiles []*github.CommitFile
	for files, response, err := prsvc.ListFiles(ctx, owner, repo, pr, options); err == nil; {
		allFiles = append(allFiles, files...)
		if response.NextPage <= options.Page {
			break
		}
		options = &github.ListOptions{Page: response.NextPage, PerPage: options.PerPage}
	}
	return allFiles
}

// GetAllReviewsComments returns all review comments of the pull request.
func GetAllReviewsComments() []*github.PullRequestComment {
	prsvc := gh.PullRequests
	options := &github.PullRequestListCommentsOptions{ListOptions: github.ListOptions{Page: 1, PerPage: 100}}

	var allComments []*github.PullRequestComment
	for comments, response, err := prsvc.ListComments(ctx, owner, repo, pr, options); err == nil; {
		allComments = append(allComments, comments...)
		if response.NextPage <= options.Page {
			break
		}
		options = &github.PullRequestListCommentsOptions{
			ListOptions: github.ListOptions{Page: response.NextPage, PerPage: options.PerPage},
		}
	}
	return allComments
}

func IsGHA() bool {
	for _, key := range requiredEnvVars {
		if val := os.Getenv(key); val == "" {
			return false
		}
	}
	return true
}

func IsPR() bool {
	return os.Getenv("GITHUB_EVENT_NAME") == "pull_request"
}

// TODO add fixing guide
func Markdown(result *header2.Result) string {
	return fmt.Sprintf(`
<!-- %s -->
[license-eye](https://github.com/apache/skywalking-eyes/tree/main/license-eye) has totally checked %d files.
| Valid | Invalid | Ignored | Fixed |
| --- | --- | --- | --- |
| %d | %d | %d | %d |
<details>
  <summary>Click to see the invalid file list</summary>

  %v
</details>
`,
		Identification,
		len(result.Success)+len(result.Failure)+len(result.Ignored),
		len(result.Success),
		len(result.Failure),
		len(result.Ignored),
		len(result.Fixed),
		"- "+strings.Join(result.Failure, "\n- "),
	)
}

type Event struct {
	PR github.PullRequest `json:"pull_request"`
}

func GetSha() (string, error) {
	filepath := os.Getenv("GITHUB_EVENT_PATH")
	logger.Log.Debugln("GITHUB_EVENT_PATH: ", filepath)
	if filepath == "" {
		return "", fmt.Errorf("failed to get event path")
	}
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	logger.Log.Debugln(filepath, "content:", string(content))

	var event Event
	if err := json.Unmarshal(content, &event); err != nil {
		return "", err
	}
	return *event.PR.Head.SHA, nil
}
