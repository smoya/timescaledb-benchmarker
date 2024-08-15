# Internal documentation and decisions
This app UI is mainly a CLI.
The main executable is found in [cmd/benchmarker/main.go](../cmd/benchmarker/main.go), and it spins up a complete CLI handled by [urfave/cli](https://github.com/urfave/cli).
Using [urfave/cli](https://github.com/urfave/cli) instead of [Cobra](https://github.com/spf13/cobra) or [Kong](https://github.com/alecthomas/kong) because it is lighter and gives us all the necessary features. No big deal if we prefer other CLI frameworks (or even using none), migration should be pretty much straightforward.

At this point, only one command is available (`benchmark queries select`), however that naming strategy has a reason: to allow other commands and subcommands to be added without breaking the interface.

More info about its arguments and options is available in the docs for each of the commands in [cmd](cmd).

## Design
The app business logic is isolated into the right services, i.e. [pkg/benchmark](../pkg/benchmark), and [pkg/timescaledb](../pkg/timescaledb) so it can be reused by other ports such as an HTTP API; allowing the app to be run by different kind of clients. 
Additionally, the code allows to create new DB implementations so it is not really limited to TimescaleDB instances.

The app follows the [Twelve-factor App](https://12factor.net/) methodology almost in its wholeness.

## Workers
The app spin ups [Workers](../pkg/query/worker.go) through a [Worker Pool](../pkg/query/pool.go).
Workers are not functions but structs under an interface. The reason for that is based on the complexity of the exercise requiring the sharding; leading to store state (like caching worker assignments, etc). Structs are a good solution **for this case** rather than dealing with detached variables declared somewhere magically.

Usually, the Worker Pool pattern assigns jobs randomly on each worker. However, as per requirements, an implementation of the [Worker Pool](../pkg/query/pool.go) has been created for sharding the jobs (queries in our case).
Workers in the worker pool are created in advance and not under demand. The reason is simplicity. However, If resources are so critical that having few workers running (are light goroutines) are not trivial, we could spin those workers dynamically under demand during the hashing moment.

### ShardedWorkerPool
Found in [pkg/pkg/query/pool.go](../pkg/query/pool.go), this implementation uses standard sharding in order to distribute queries across workers. 

The chosen algorithm, in this case, has been [FNV-1A](https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function). I use it for calculating the modulo of the hash and total number of workers in order to determine the worker to use. 
I decided to use FNV-1A because it is included in Go’s stdlib, it is fast, it is more resistant to collisions than most algorithms and randomness distribution is pretty good. 
An improvement would be to use Murmur 3 as an alternative as it seems to be a bit faster and improves randomness distribution, however, you will need an external dependency. 

See [this Stack Exchange thread](https://softwareengineering.stackexchange.com/a/145633/450177) to see an amazing comparison of the different algorithms.

## Stats
I decided to calculate all stats myself instead of using third party dependencies. 
By doing that, I properly understood the expectations of what the values were supposed to be. Of course, now that there are tests for this part, switching to a third party library will be just straightforward and viable. 

## Test code coverage
I didn't find having a 100% of code coverage but I must admit I would increase it in some parts of the code as the [cmd/benchmarker/cmd/select.go](../cmd/benchmarker/cmd/select.go) file. 
That would require to improve the abstractions and to create others such as one for reading from STDOUT and check that the results are properly printed, etc. I considered this a lot of work for this very first version.

## benchmark queries select STDIN support
I allowed reading the CSV file by specifying the file path or rather by piping the file through STDIN. 
I disabled STDIN in terminal mode (manual introduction of data) since IMHO doesn’t make sense. 
By disabling this, we give a better DX since the command doesn’t wait endlessly in the case of no file path and no file piped (terminal).

## External dependencies
- Using [logrus](https://github.com/sirupsen/logrus) and [testify](https://github.com/stretchr/testify) in order to speed up in development time. Not a big fan because several reasons, but I have to admit it speeds up development on brand new code. 
- Using [github.com/pashagolub/pgxmock](https://github.com/pashagolub/pgxmock/) for mocking the DB. 
  - I believe most of mocking libraries are not needed if you have a good segregation on interfaces. 
  - However, there are exceptions with external dependencies, where sometimes is better to rely on their mock packages or third parties because the big amount of code needed for mocking them. 
  - Is not worth IMHO spending more time on it atm.
- Using [github.com/jedib0t/go-pretty](https://github.com/jedib0t/go-pretty) in order to pretty print the results of the `benchmark queries select` command. 
  - It might seem an unnecessary dependency but it offers a really good addition besides the good appeal of the results: it allows printing in other formats such as Markdown, CSV, HTML...
  - Since it also does sorting and grouping of the **final** results, it is not an operation you can do concurrently. That has a cons, and it is the fact you need to store all query results into an array (consuming memory) so you can sort and pretty print later. 
  - If efficiency and speed really (really) matter, then I would just remove such a pretty print feature in favor of printing each result when read from the results channel (in a goroutine), again losing all sorting and nice printing.

## Other near-future improvements
- To support several CSV types, by specifying the fields to use of the CSV.
- To support other inputs besides CSV.
- To add support for other DBs.
- Rate limit for query runners.
- Stats collector to be a real collector, isolated from the query package, with representation of histograms and its metrics.
- Add performance tests
- Track anonymous usage metrics
- To create its own Helm chart so it can be deployed into K8s, just as I did in, for example: https://github.com/asyncapi-archived-repos/event-gateway/tree/master/deployments/k8s
- Code:
    - At this moment, the query runner is tied to pgx PostgreSQL driver. Even though the Query interface is agnostic, I needed to “hardcode” the query results scans variables (TS, max, min). In order to make this agnostic, we would need to additionally extract the logic that process (after query) the rows. I decided to keep it like it is now for simplicity and time constraints.
    - Query interface splitting statement and Args.
    - DB Factory based on URIs, not only support for TimescaleDB.