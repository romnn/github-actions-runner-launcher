## github-actions-runner-launcher

[![Build Status](https://travis-ci.com/romnnn/github_actions_runner_launcher.svg?branch=master)](https://travis-ci.com/romnnn/github_actions_runner_launcher)
[![GitHub](https://img.shields.io/github/license/romnnn/github_actions_runner_launcher)](https://github.com/romnnn/github-actions-runner-launcher)
[![GoDoc](https://godoc.org/github.com/romnnn/github-actions-runner-launcher?status.svg)](https://godoc.org/github.com/romnnn/github-actions-runner-launcher) [![Docker Pulls](https://img.shields.io/docker/pulls/romnn/github_actions_runner_launcher)](https://hub.docker.com/r/romnn/github_actions_runner_launcher) [![Test Coverage](https://codecov.io/gh/romnnn/github_actions_runner_launcher/branch/master/graph/badge.svg)](https://codecov.io/gh/romnnn/github_actions_runner_launcher)
[![Release](https://img.shields.io/github/release/romnnn/github_actions_runner_launcher)](https://github.com/romnnn/github-actions-runner-launcher/releases/latest)

Automatic setup and registration of GitHub actions runner instances. If you have tried [github.com/myoung34/docker-github-actions-runner](https://github.com/myoung34/docker-github-actions-runner) or similiar options for running the action runners as docker containers for the convenience, but want to run them on bare-metal, this is just right for you!

Configuration works just like with docker-compose as shown in this `sample.yml`:
```yaml
version: '3.3'

services:
  my-runner:
    environment:
      REPO_URL: https://github.com/my/repo
      # Use access token to automatically obtain runner tokens from the github API
      ACCESS_TOKEN: <MY-SECRET-GITHUB-ACCESS-TOKEN>
      # The runner token must be specified otherwise
      RUNNER_TOKEN: <MY-SECRET-RUNNER-TOKEN>
      RUNNER_NAME: my-runner
      RUNNER_WORKDIR: /my/runners/work/dir
      ORG_RUNNER: "false"
      ORG_NAME: my-org
      LABELS: linux,x64
```

You can then start with
```bash
go get github.com/romnnn/github-actions-runner-launcher/cmd/github-actions-runner-launcher --config sample.yml
```

You can also download pre built binaries from the [releases page](https://github.com/romnnn/github-actions-runner-launcher/releases), or use the `docker` image:

```bash
docker pull romnn/github-actions-runner-launcher
```

For a list of options, run with `--help`.


#### Usage as a library

```golang
import "github.com/romnnn/github-actions-runner-launcher"
```

For more examples, see `examples/`.


#### Development

######  Prerequisites

Before you get started, make sure you have installed the following tools::

    $ python3 -m pip install -U cookiecutter>=1.4.0
    $ python3 -m pip install pre-commit bump2version invoke ruamel.yaml halo
    $ go get -u golang.org/x/tools/cmd/goimports
    $ go get -u golang.org/x/lint/golint
    $ go get -u github.com/fzipp/gocyclo
    $ go get -u github.com/mitchellh/gox  # if you want to test building on different architectures

**Remember**: To be able to excecute the tools downloaded with `go get`, 
make sure to include `$GOPATH/bin` in your `$PATH`.
If `echo $GOPATH` does not give you a path make sure to run
(`export GOPATH="$HOME/go"` to set it). In order for your changes to persist, 
do not forget to add these to your shells `.bashrc`.

With the tools in place, it is strongly advised to install the git commit hooks to make sure checks are passing in CI:
```bash
invoke install-hooks
```

You can check if all checks pass at any time:
```bash
invoke pre-commit
```

Note for Maintainers: After merging changes, tag your commits with a new version and push to GitHub to create a release:
```bash
bump2version (major | minor | patch)
git push --follow-tags
```

#### Note

This project is still in the alpha stage and should not be considered production ready.
