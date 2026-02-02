# labeler-action

[![Build Status](https://github.com/jimschubert/labeler-action/actions/workflows/build.yml/badge.svg)](https://github.com/jimschubert/labeler-action/actions/workflows/build.yml) 
[![Go Version](https://img.shields.io/github/go-mod/go-version/jimschubert/labeler-action)](https://github.com/jimschubert/labeler-action/blob/master/go.mod)
[![License](https://img.shields.io/github/license/jimschubert/labeler-action)](https://github.com/jimschubert/labeler-action/blob/master/LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/jimschubert/labeler-action)](https://github.com/jimschubert/labeler-action/releases/latest)
[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-Labeler-blue?logo=github)](https://github.com/marketplace/actions/auto-labeler)

GitHub Action allowing for applying labels to issues and pull requests based on patterns found in the title or description.

## Deprecation Notice

⚠️ **The `GITHUB_TOKEN` input is deprecated and will be removed in v3.** Pass the token via `env` instead:

```yaml
- uses: jimschubert/labeler-action@v2
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Usage

### Define `.github/labeler.yml`

This action requires a configuration file defined at `.github/labeler.yml` in your repository. The contents must follow either the *simple* schema or the *full* schema shown below.

**NOTE** The file must exist on your release branch (e.g. `master`, `main`, etc.).

Feel free to use one of the following schema examples to get started.

#### Simple Schema

```yaml
# labeler "simple" schema
# Comment is applied to both issues and pull requests.
# If you need a more robust solution, consider the "full" schema.
comment: |
  👍 Thanks for this!
  🏷 I have applied any labels matching special text in your issue.

  Please review the labels and make any necessary changes.

# Labels is an object where:
# - keys are labels
# - values are array of string patterns to match against title + body in issues/prs
labels:
  'bug':
    - '\bbug[s]?\b'
  'help wanted':
    - '\bhelp( wanted)?\b'
  'duplicate':
    - '\bduplicate\b'
    - '\bdupe\b'
  'enhancement':
    - '\benhancement\b'
  'question':
    - '\bquestion\b'
```

#### Full Schema

```yaml
# labeler "full" schema

# enable labeler on issues, prs, or both.
enable:
  issues: true
  prs: true
# comments object allows you to specify a different message for issues and prs

comments:
  issues: |
    Thanks for opening this issue!
    I have applied any labels matching special text in your title and description.

    Please review the labels and make any necessary changes.
  prs: |
    Thanks for the contribution!
    I have applied any labels matching special text in your title and description.

    Please review the labels and make any necessary changes.

# Labels is an object where:
# - keys are labels
# - values are objects of { include: [ pattern ], exclude: [ pattern ] }
#    - pattern must be a valid regex, and is applied globally to
#      title + description of issues and/or prs (see enabled config above)
#    - 'include' patterns will associate a label if any of these patterns match
#    - 'exclude' patterns will ignore this label if any of these patterns match
#    - 'branches' is an optional array of branch names (or patterns) to limit labeling according to PR target branch
labels:
  'bug':
    include:
      - '\bbug[s]?\b'
    exclude: []
    branches:
      - 'master'
      - 'main'
  'help wanted':
    include:
      - '\bhelp( me)?\b'
    exclude:
      - '\b\[test(ing)?\]\b'
  'enhancement':
    include:
      - '\bfeat\b'
    exclude: []

```

### Create a Workflow

The action requires the `GITHUB_TOKEN` environment variable. This token allows the action to access the GitHub API for your account. Workflows automatically provide a default `GITHUB_TOKEN`, which provides full API access. You can create a secret from a [new token](https://github.com/settings/tokens) with `public_repo` scope to limit the action's footprint.
 
**NOTE** Binding to issue or pull_request `edit` actions is _not_ recommended.

Create a workflow definition, for example `.github/workflows/community.yml`:

```yaml
name: Community
on: 
  issues:
    types: [opened, edited, milestoned]
  pull_request_target:
    types: [opened]

permissions:
  issues: write
  pull-requests: write

jobs:

  labeler:
    runs-on: ubuntu-latest

    steps:
    - name: Check Labels
      id: labeler
      uses: jimschubert/labeler-action@v2
      env:
        GITHUB_TOKEN: ${{secrets.GITHUB_TOKEN}}
```

Notice that the PR event is [pull_request_target](https://docs.github.com/en/actions/reference/events-that-trigger-workflows#pull_request_target) rather than `pull_request`. The difference is that `pull_request_target` is a newer event which runs in the context of the base repository, allowing access to the secrets necessary to label the pull request.
If you use `pull_request`, the action will be able to label pull requests originating from the same repository (i.e. no forks!).

## License

This project is licensed under Apache 2.0
