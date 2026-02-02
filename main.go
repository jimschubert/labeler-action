package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v29/github"
	"github.com/jimschubert/labeler"
	act "github.com/sethvargo/go-githubactions"
	log "github.com/sirupsen/logrus"
)

var (
	version = "dev"
	commit  = "none"
)

type executor interface {
	Execute() error
}

var newLabelerWithOptions = func(opts ...labeler.OptFn) (executor, error) {
	l, err := labeler.NewWithOptions(opts...)
	return l, err
}

func runLabelerFromEnv() error {
	fmt.Printf("Running labeler %s (%s)\n", version, commit)

	eventName := os.Getenv("GITHUB_EVENT_NAME")

	githubToken := act.GetInput("GITHUB_TOKEN")
	if githubToken == "" {
		var ok bool
		// allow for local testing
		githubToken, ok = os.LookupEnv("GITHUB_TOKEN")
		if !ok || githubToken == "" {
			return fmt.Errorf("missing environment variable 'GITHUB_TOKEN' in labeler action configuration")
		}
	} else {
		fmt.Println("::warning::The GITHUB_TOKEN input is deprecated and will be removed in v3. Pass it via env instead. See docs for details.")
	}

	// actions will pass input GITHUB_TOKEN as env INPUT_GITHUB_TOKEN, so set this back to GITHUB_TOKEN for the lib.
	_ = os.Setenv("GITHUB_TOKEN", githubToken)
	_ = os.Setenv("LOG_LEVEL", "info")

	event, err := os.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil || event == nil {
		return fmt.Errorf("can't read events: %w", err)
	}

	re := regexp.MustCompile(`\r?\n\s*`)
	event = re.ReplaceAll(event, []byte(""))

	var id int
	switch eventName {
	case "issues":
		var issue github.IssuesEvent
		err = json.Unmarshal(event, &issue)
		if err != nil {
			return fmt.Errorf("can't unmarshal json: %w", err)
		}
		if issue.Issue != nil {
			id = (*issue.Issue).GetNumber()
		}
	case "pull_request", "pull_request_target":
		var pr *github.PullRequestEvent
		err = json.Unmarshal(event, &pr)
		if err != nil {
			return fmt.Errorf("can't unmarshal json: %w", err)
		}
		if pr.PullRequest != nil {
			id = (*pr.PullRequest).GetNumber()
		}
	}

	data := string(event)
	log.WithFields(log.Fields{"data": data, "type": eventName}).Debug("Processing event")

	// GITHUB_REPOSITORY_OWNER is only recently introduced in the runner
	// see https://github.com/actions/runner/pull/378
	// GITHUB_REPOSITORY_OWNER would be used by forks
	// GITHUB_ACTOR would be used by sources where GITHUB_REPOSITORY_OWNER may not exist
	owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	if owner == "" {
		owner = os.Getenv("GITHUB_ACTOR")
	}

	repo := os.Getenv("GITHUB_REPOSITORY")
	repoParts := strings.Split(repo, "/")
	repo = repoParts[len(repoParts)-1]

	labelOpts := make([]labeler.OptFn, 0)
	labelOpts = append(labelOpts, labeler.WithOwner(owner))
	labelOpts = append(labelOpts, labeler.WithRepo(repo))
	labelOpts = append(labelOpts, labeler.WithEvent(eventName))
	labelOpts = append(labelOpts, labeler.WithID(id))
	labelOpts = append(labelOpts, labeler.WithData(data))

	if configPath := act.GetInput("config_path"); configPath != "" {
		labelOpts = append(labelOpts, labeler.WithConfigPath(configPath))
	}

	l, err := newLabelerWithOptions(labelOpts...)
	if err != nil {
		return fmt.Errorf("could not construct a labeler: %w", err)
	}

	err = l.Execute()
	if err != nil {
		return fmt.Errorf("failed to execute: %w", err)
	}

	log.Info("Done labeling.")
	return nil
}

func main() {
	if err := runLabelerFromEnv(); err != nil {
		log.Fatal(err)
	}
}
