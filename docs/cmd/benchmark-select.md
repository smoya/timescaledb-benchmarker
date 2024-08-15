# benchmark select
`benchmark select` CMD benchmarks SELECT queries. The queries should be in CSV format and contain 3 headers:

- `hostname`: string representation of the hostname.
- `start_time`: date time following the format `<year>-<month>-<day>` representing the start time of the date range used in queries.
- `end_time`: date time following the format `<year>-<month>-<day>` representing the end time of the date range used in queries.

Example:
```csv
hostname,start_time,end_time
host_a,2017-12-31 08:59:22,2017-01-01 09:59:22
```

## Usage
```shell
timescaledb_benchmarker benchmark select [command options]
```

## Config
| Flag            | Alias | Env var                                         | Description                                                                                                                                             | format                      | Required | Default     | Example                                                            |
|-----------------|-------|-------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------|----------|-------------|--------------------------------------------------------------------|
| --file          | -f    | TIMESCALEDB_BENCHMARKER_BENCHMARK_FILE          | Path to a csv file containing raw query fields (format as specified above)                                                                              | file path                   | No       | STDIN input | -f /data/query_params.csv                                          |
| --db_uri        |       | TIMESCALEDB_BENCHMARKER_BENCHMARK_DB_URI        | TimescaleDB Connection URI                                                                                                                              | Postgres Conn URI           | Yes      |             | --db_uri postgres://username:password@localhost:5432/database_name |
| --workers       | -w    | TIMESCALEDB_BENCHMARKER_BENCHMARK_WORKERS       | Number of query workers executing Queries concurrently. Different from Postgress pool size, which can be configured in parallel through the db_uri flag | uint                        | No       | 5           | -w 10                                                              |
| --timeout       | -t    | TIMESCALEDB_BENCHMARKER_BENCHMARK_TIMEOUT       | Timeout for each query. A string with is a sequence of decimal numbers, each with optional fraction and a unit suffix such as `300ms`, or `2h45m`       | Duration as string          | No       | 200ms       | -t 400ms                                                           |
| --debug         | -d    | TIMESCALEDB_BENCHMARKER_BENCHMARK_DEBUG         | Debug mode. Enable it for printing debug logs                                                                                                           | boolean                     | No       | false       | -d true                                                            |
| --output_format |       | TIMESCALEDB_BENCHMARKER_BENCHMARK_OUTPUT_FORMAT | Output print format. By default, human readable output for printing in the console                                                                      | enum[human,csv,tsv,md,html] | No       | human       | --output-format md                                                 |

### Reference
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