# Timescale Benchmarker
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->
[![License](https://img.shields.io/github/license/smoya/timescaledb-benchmarker)](https://github.com/smoya/timescaledb-benchmarker/blob/master/LICENSE)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/smoya/timescaledb-benchmarker/.github%2Fworkflows%2Frelease.yml)
[![last commit](https://img.shields.io/github/last-commit/smoya/timescaledb-benchmarker)](https://github.com/smoya/timescaledb-benchmarker/commits/master)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/smoya/timescaledb-benchmarker)
[![All Contributors](https://img.shields.io/github/all-contributors/smoya/timescaledb-benchmarker?color=ee8449&style=flat-square)](#contributors)


A Benchmarker for TimescaleDB select queries. This is just a demo project and is not meant for production use.

- [Installation](#installation)
- [Demo](#demo)
- [Usage](#usage)
- [Versioning and maintenance](#versioning-and-maintenance)
- [Development](#development)
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
This CLI app is split into several commands (at this moment, just one). 

### benchmark select
Benchmarks SELECT queries. The queries should be in CSV format and contain 3 headers:

- `hostname`: string representation of the hostname.
- `start_time`: date time following the format `<year>-<month>-<day>` representing the start time of the date range used in queries.
- `end_time`: date time following the format `<year>-<month>-<day>` representing the end time of the date range used in queries.

Example:
```csv
hostname,start_time,end_time
host_a,2017-12-31 08:59:22,2017-01-01 09:59:22
```

#### Usage 
```shell
timescaledb_benchmarker benchmark select [command options]
```

#### Config
| Flag            | Alias | Env var                                         | Description                                                                                                                                             | format                      | Required | Default     | Example                                                            |
|-----------------|-------|-------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------|----------|-------------|--------------------------------------------------------------------|
| --file          | -f    | TIMESCALEDB_BENCHMARKER_BENCHMARK_FILE          | Path to a csv file containing raw query fields (format as specified above)                                                                              | file path                   | No       | STDIN input | -f /data/query_params.csv                                          |
| --db_uri        |       | TIMESCALEDB_BENCHMARKER_BENCHMARK_DB_URI        | TimescaleDB Connection URI                                                                                                                              | Postgres Conn URI           | Yes      |             | --db_uri postgres://username:password@localhost:5432/database_name |
| --workers       | -w    | TIMESCALEDB_BENCHMARKER_BENCHMARK_WORKERS       | Number of query workers executing Queries concurrently. Different from Postgress pool size, which can be configured in parallel through the db_uri flag | uint                        | No       | 5           | -w 10                                                              |
| --timeout       | -t    | TIMESCALEDB_BENCHMARKER_BENCHMARK_TIMEOUT       | Timeout for each query. A string with is a sequence of decimal numbers, each with optional fraction and a unit suffix such as `300ms`, or `2h45m`       | Duration as string          | No       | 200ms       | -t 400ms                                                           |
| --debug         | -d    | TIMESCALEDB_BENCHMARKER_BENCHMARK_DEBUG         | Debug mode. Enable it for printing debug logs                                                                                                           | boolean                     | No       | false       | -d true                                                            |
| --output_format |       | TIMESCALEDB_BENCHMARKER_BENCHMARK_OUTPUT_FORMAT | Output print format. By default, human readable output for printing in the console                                                                      | enum[human,csv,tsv,md,html] | No       | human       | --output-format md                                                 |

#### Example
```shell
timescaledb_benchmarker benchmark select -f /data/query_params.csv --db_uri postgres://username:password@localhost:5432/database_name
```

> [!TIP]
> You can redirect or pipe any CSV (without specifying the --file or -f flags). 
> Useful for many use cases, as requesting the file from the Internet, from a result of a command, etc.
> 
> Examples:
> - `timescaledb_benchmarker benchmark select --db-uri ${DB_URI} < /data/query_params.csv`
> - `curl -s https://example.com/csv_file.csv | timescaledb_benchmarker benchmark select --db-uri ${DB_URI}`


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