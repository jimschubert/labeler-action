// Copyright 2020 Jim Schubert
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v29/github"
	"github.com/jimschubert/labeler"
	act "github.com/sethvargo/go-githubactions"
	log "github.com/sirupsen/logrus"
)

func main() {
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	githubToken := act.GetInput("GITHUB_TOKEN")
	if githubToken == "" {
		var ok bool
		// allow for local testing
		githubToken, ok = os.LookupEnv("GITHUB_TOKEN")
		if !ok {
			log.Fatal("Missing input 'GITHUB_TOKEN' in labeler action configuration.")
		}
	}

	// github actions will pass input GITHUB_TOKEN as env INPUT_GITHUB_TOKEN, so set this back to GITHUB_TOKEN for the lib.
	_ = os.Setenv("GITHUB_TOKEN", githubToken)
	_ = os.Setenv("LOG_LEVEL", "info")

	event, err := ioutil.ReadFile(os.Getenv("GITHUB_EVENT_PATH"))
	if err != nil {
		log.Fatalf("Can't read events: %s", err)
	}

	re := regexp.MustCompile(`\r?\n\s*`)
	event = re.ReplaceAll(event,[]byte(""))

	var id int
	switch eventName {
	case "issues":
		var issue github.IssuesEvent
		err = json.Unmarshal(event, &issue)
		if err != nil {
			log.Fatalf("Can't unmarshal json: %s", err)
		}

		id = (*issue.Issue).GetNumber()
	case "pull_request":
		var pr *github.PullRequestEvent
		err = json.Unmarshal(event, &pr)
		if err != nil {
			log.Fatalf("Can't unmarshal json: %s", err)
		}
		id = (*pr.PullRequest).GetNumber()
	}

	data := string(event)

	log.WithFields(log.Fields{"data": data, "type": eventName}).Debug("Processing event")

	owner := os.Getenv("GITHUB_ACTOR")
	repo := os.Getenv("GITHUB_REPOSITORY")
	repo = strings.Replace(repo, owner+ "/", "", 1)
	l, err := labeler.New(owner, repo, eventName, id, &data)
	if err != nil {
		log.Fatalf("Could not construct a labeler %s", err)
	}

	err = l.Execute()
	if err != nil {
		log.Fatalf("Failed to execute %v", err)
	}

	log.Info("Done labeling.")
}