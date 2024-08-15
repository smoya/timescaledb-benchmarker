# Timescale Benchmarker

[![License](https://img.shields.io/github/license/smoya/timescaledb-benchmarker)](https://github.com/smoya/timescaledb-benchmarker/blob/master/LICENSE)
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/smoya/timescaledb-benchmarker/.github%2Fworkflows%2Frelease.yml)
[![last commit](https://img.shields.io/github/last-commit/smoya/timescaledb-benchmarker)](https://github.com/smoya/timescaledb-benchmarker/commits/master)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/smoya/timescaledb-benchmarker)

A Benchmarker for your TimescaleDB instance. This is just a demo project and is not meant for production use.

- [Installation](#installation)
- [Demo](#demo)
- [Usage](#usage)
- [Development](#development)
- [Internal docs](#internal-docs)
- [Contributing](#contributing)
- [Contributors](#contributors)

## Installation

### Go install
You can install the binary globally by running:

```shell
go install github.com/smoya/timescaledb-benchmarker/cmd/benchmarker@latest
```

### Precompiled binaries
You will find precompiled binaries on the [releases page](https://github.com/smoya/timescaledb-benchmarker/releases/). Architectures supported:

- Darwin AMD64 (Apple Intel)
- Darwin ARM64 (Apple Silicon)
- Linux AMD64
- Windows AMD64

Please feel free to open an issue requesting the inclusion of any other architecture compiled binary for new releases.

### Docker
This project contains a [Dockerfile](Dockerfile) in order to build its Docker image. 
You can find all the Docker images at [Docker Hub](https://hub.docker.com/repository/docker/smoya/timescaledb-benchmarker).

Run this CLI app from : 
```shell
docker run smoya/timescaledb-benchmarker
```

Please refer to the [Usage](#usage) section for the full list of available commands.

## Demo
This project contains a [Docker Compose file](deployments/docker-compose/compose.yml) that can be used for demoing the benchmarker. Fixtures are available under [deployments/data](deployments/data). 

Please run:
```shell
docker compose -f deployments/docker-compose/compose.yml up
```

This will spin up two containers:
1. `timescaledb`: using [timescale/timescaledb](https://hub.docker.com/r/timescale/timescaledb) docker image. This spins up a TimescaleDB database (no HA). It is configured to run init scripts that will create and populate the DB with the files located under [deployments/data](deployments/data).
2. `timescaledb-benchmarker`: using [smoya/timescaledb-benchmarker], which is the image that we build in this repository.

You can overwrite the default environment variables by using the [--env](https://docs.docker.com/compose/environment-variables/set-environment-variables/#set-environment-variables-with-docker-compose-run---env) flag. Please refer to [Configuration](#configuration) for the full list of supported flags and environment variables.

### Executing the benchmark
The benchmarker will use the data from [deployments/data/query_params.csv](deployments/data/query_params.csv) in order to generate the queries for the benchmark.

Please run: 
```shell
docker exec docker-compose-timescaledb-benchmarker-1 timescaledb-benchmarker benchmark select
```

The [deployments/docker-compose/compose.yml](deployments/docker-compose/compose.yml) file contains the required environment variables set. However, you can overwrite them or add new ones via the `--env` flag.

I.e., if you want to disable the debug log level, please run:
```shell
docker exec --env TIMESCALEDB_BENCHMARKER_BENCHMARK_DEBUG=false docker-compose-timescaledb-benchmarker-1 timescaledb-benchmarker benchmark select
```

## Usage
This CLI app is a binary that offers mainly a CLI GUI, split into several commands.
Find the docs for each of the commands in [docs/cmd](docs/cmd).

## Development

### Prerequisites
- [Go v1.23](https://go.dev/dl/) or higher. 
  - If you are in `v1.22`,  I recommend enabling the [rangefunc experiment](https://go.dev/wiki/RangefuncExperiment).
- Optional: Docker. Used for building and running the CLI app in an isolated env.
  - Additionally, there is a [compose.yml](deployments/docker-compose/compose.yml) that spins up this app plus a TimescaleDB DB instance (including fixtures). More info on the [Demo](#demo) section.

#### Build
A dedicated [Makefile](Makefile) target is available:

```shell
make build
```

You will find the compiled binary under [bin/out](bin/out) directory.

#### Lint
[golangci-lint](https://golangci-lint.run/) is used for linting the code. A dedicated [Makefile](Makefile) target is available:

```shell
make lint
```

### Test
Unit tests are available for each .go file. [Testify](github.com/stretchr/testify) toolkit is installed and in use.

There are **two** [Makefile](Makefile) targets:

1. `make test`, which runs all tests with the native go race condition detector.
2. `make coverage`, which runs all tests as `make test` does plus it shows up code coverage statistics.

### Releases
Releases are handled automatically on each push to the `main` branch via [Semantic Release](https://semantic-release.gitbook.io/semantic-release).
Commits are following [Conventional Commits](https://www.conventionalcommits.org/), and releases will only happen when commits have certain prefixes:
- `feat`: a [MINOR](http://semver.org/#summary) release will happen. Example: `feat: add a new feature`.
- `fix`: a [PATCH](http://semver.org/#summary) release will happen. Example: `fix: add a new fix`.
- any of the previous with a `!` suffix: a [MAJOR](http://semver.org/#summary) release will happen. Example: `feat!: add a new breaking change`.

## Internal docs
A set of internal docs and design decisions can be found in [docs/internal.md](docs/internal.md). If you want to know why some things are like they are, that is for you.

## Contributing
Contributions are always encouraged! Please submit your PR's or issues. Tag or assign any of the [code owners](CODEOWNERS) (at this moment, only me: @smoya) as reviewers.
Be aware of the [LICENSE](LICENSE) of this project before submitting your PR.

Feel free to ping me if you have any doubts or need an onboarding of this project.

## Contributors
<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://smoya.dev/"><img src="https://avatars.githubusercontent.com/u/1083296?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Sergio Moya</b></sub></a><br /><a href="https://github.com/smoya/timescaledb-benchmarker/commits?author=smoya" title="Code">💻</a> <a href="https://github.com/smoya/timescaledb-benchmarker/commits?author=smoya" title="Documentation">📖</a> <a href="#example-smoya" title="Examples">💡</a> <a href="#ideas-smoya" title="Ideas, Planning, & Feedback">🤔</a> <a href="#infra-smoya" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="#maintenance-smoya" title="Maintenance">🚧</a> <a href="#projectManagement-smoya" title="Project Management">📆</a> <a href="#research-smoya" title="Research">🔬</a> <a href="https://github.com/smoya/timescaledb-benchmarker/commits?author=smoya" title="Tests">⚠️</a> <a href="#tutorial-smoya" title="Tutorials">✅</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->