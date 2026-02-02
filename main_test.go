package main

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/google/go-github/v29/github"
	"github.com/jimschubert/labeler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLabeler struct {
	mock.Mock
	executor
}

func (m *MockLabeler) Execute() error {
	args := m.Called()
	return args.Error(0)
}

// Helper to set and restore env vars
func withEnv(key, value string, fn func()) {
	orig := os.Getenv(key)
	os.Setenv(key, value)
	defer os.Setenv(key, orig)
	fn()
}

func withMockLabeler(fn func(ex *MockLabeler)) {
	mockLabeler := new(MockLabeler)
	mockLabeler.On("Execute").Return(nil)

	orig := newLabelerWithOptions
	defer func() { newLabelerWithOptions = orig }()

	newLabelerWithOptions = func(...labeler.OptFn) (executor, error) {
		return mockLabeler, nil
	}

	fn(mockLabeler)
}

func withMockLabelerErr(err error, fn func(ex *MockLabeler)) {
	mockLabeler := new(MockLabeler)
	mockLabeler.On("Execute").Return(err)

	orig := newLabelerWithOptions
	defer func() { newLabelerWithOptions = orig }()

	newLabelerWithOptions = func(...labeler.OptFn) (executor, error) {
		return mockLabeler, nil
	}

	fn(mockLabeler)
}

// Example test for issues event
func TestRunLabeler_IssuesEvent(t *testing.T) {
	issue := &github.Issue{Number: github.Int(42)}
	event := github.IssuesEvent{Issue: issue}
	eventBytes, _ := json.Marshal(event)

	withEnv("GITHUB_EVENT_NAME", "issues", func() {
		withEnv("GITHUB_TOKEN", "token", func() {
			withEnv("GITHUB_EVENT_PATH", "event.json", func() {
				withEnv("GITHUB_REPOSITORY", "jimschubert/testrepo", func() {
					withEnv("GITHUB_REPOSITORY_OWNER", "jimschubert", func() {
						// Write event file
						os.WriteFile("event.json", eventBytes, 0644)
						defer os.Remove("event.json")

						withMockLabeler(func(mockLabeler *MockLabeler) {
							err := runLabelerFromEnv()
							assert.NoError(t, err)
							mockLabeler.AssertExpectations(t)
						})
					})
				})
			})
		})
	})
}

func TestRunLabeler_MissingToken(t *testing.T) {
	withEnv("GITHUB_TOKEN", "", func() {
		withEnv("GITHUB_EVENT_NAME", "issues", func() {
			withEnv("GITHUB_EVENT_PATH", "event.json", func() {
				err := runLabelerFromEnv()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "missing environment variable 'GITHUB_TOKEN' in labeler action configuration")
			})
		})
	})
}

func TestRunLabeler_CantReadConfig(t *testing.T) {
	withEnv("GITHUB_EVENT_NAME", "issues", func() {
		withEnv("GITHUB_TOKEN", "token", func() {
			withEnv("GITHUB_EVENT_PATH", "$.json", func() {
				withEnv("GITHUB_REPOSITORY", "jimschubert/testrepo", func() {
					withEnv("GITHUB_REPOSITORY_OWNER", "jimschubert", func() {
						err := runLabelerFromEnv()
						assert.Error(t, err)
						assert.Contains(t, err.Error(), "can't read events: open $.json: no such file or directory")
					})
				})
			})
		})
	})
}

func TestRunLabeler_InvokesLabelerSuccess(t *testing.T) {
	withEnv("GITHUB_EVENT_NAME", "issues", func() {
		withEnv("GITHUB_TOKEN", "token", func() {
			withEnv("GITHUB_EVENT_PATH", "testdata/issue.json", func() {
				withEnv("GITHUB_REPOSITORY", "jimschubert/testrepo", func() {
					withEnv("GITHUB_REPOSITORY_OWNER", "jimschubert", func() {

						withMockLabeler(func(mockLabeler *MockLabeler) {
							err := runLabelerFromEnv()
							assert.NoError(t, err)
						})
					})
				})
			})
		})
	})
}

func TestRunLabeler_InvokesLabelerError(t *testing.T) {
	withEnv("GITHUB_EVENT_NAME", "issues", func() {
		withEnv("GITHUB_TOKEN", "token", func() {
			withEnv("GITHUB_EVENT_PATH", "testdata/issue.json", func() {
				withEnv("GITHUB_REPOSITORY", "jimschubert/testrepo", func() {
					withEnv("GITHUB_REPOSITORY_OWNER", "jimschubert", func() {

						withMockLabelerErr(errors.New("some error"), func(mockLabeler *MockLabeler) {
							err := runLabelerFromEnv()
							assert.Error(t, err)
							assert.Contains(t, err.Error(), "some error")
						})
					})
				})
			})
		})
	})
}

func TestRunLabeler_DeprecationWarning(t *testing.T) {
	// Capture stdout to verify deprecation warning
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// INPUT_GITHUB_TOKEN simulates passing token via 'with:' in actions
	withEnv("INPUT_GITHUB_TOKEN", "token", func() {
		withEnv("GITHUB_EVENT_NAME", "issues", func() {
			withEnv("GITHUB_EVENT_PATH", "testdata/issue.json", func() {
				withEnv("GITHUB_REPOSITORY", "jimschubert/testrepo", func() {
					withEnv("GITHUB_REPOSITORY_OWNER", "jimschubert", func() {
						withMockLabeler(func(mockLabeler *MockLabeler) {
							err := runLabelerFromEnv()
							assert.NoError(t, err)
						})
					})
				})
			})
		})
	})

	w.Close()
	os.Stdout = oldStdout

	var buf [1024]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	assert.Contains(t, output, "::warning::The GITHUB_TOKEN input is deprecated and will be removed in v3. Pass it via env instead. See docs for details.")
}

