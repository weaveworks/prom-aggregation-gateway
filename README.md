# Prometheus Aggregation Gateway

Prometheus Aggregation Gateway is a push gateway that aggregates metrics for Prometheus.  As opposed to the official [Prometheus Pushgateway](https://github.com/prometheus/pushgateway), this service aggregates the sample values it receives.

* Counters where all labels match are added up.
* Histograms are added up; if bucket boundaries are mismatched then the result has the union of all buckets and counts are given to the lowest bucket that fits.
* Gauges are also added up (but this may not make any sense)
* Summaries are discarded.

## How to use

Send metrics in [Prometheus format](https://prometheus.io/docs/instrumenting/exposition_formats/) to `/metrics/`

E.g. if you have the program running locally:

```bash
echo 'http_requests_total{method="post",code="200"} 1027' | curl --data-binary @- http://localhost/metrics/
```

You can also add labels that will apply to all metrics that are passed. For example, the following will add the `domain="sometest.com"` and `instance="nginx-1"` labels to both metrics:

```bash
echo '
http_requests_total{method="post",code="200"} 1027
http_errors_total{method="post",code="500"} 6
' | curl --data-binary @- http://localhost/metrics/domain/sometest.com/instance/nginx-1
```

Now you can push your metrics using your favorite Prometheus client.

E.g. in Python using [prometheus/client_python](https://github.com/prometheus/client_python):

```python
from prometheus_client import CollectorRegistry, Counter, push_to_gateway
registry = CollectorRegistry()
counter = Counter('some_counter', "A counter", registry=registry)
counter.inc()
push_to_gateway('localhost', job='my_job_name', registry=registry)
```

Then have your Prometheus scrape metrics at `/metrics`.

### Running the service


help command can be useful when calling the binary to see what commands and flags are available

```bash
$ prom-aggregation-gateway --help

prometheus aggregation gateway

Usage:
  prom-aggregation-gateway [flags]
  prom-aggregation-gateway [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  start       starts up the server
  version     Show version information

Flags:
      --AuthUsers strings        List of allowed auth users and their passwords comma separated
                                  Example: "user1=pass1,user2=pass2"
      --apiListen string         Listen for API requests on this host/port. (default ":80")
      --cors string              The 'Access-Control-Allow-Origin' value to be returned. (default "*")
  -h, --help                     help for prom-aggregation-gateway
      --lifecycleListen string   Listen for lifecycle requests (health, metrics) on this host/port (default ":8888")

Use "prom-aggregation-gateway [command] --help" for more information about a command.
```

Any flags you see above can also be set by `ENV_VARIABLES`. ENV_VARS must have a prefix of `PAG_`, for example `PAG_AUTHUSERS=user1=pass1,user2=pass2` will start the service with basic auth. If an ENV_VARIABLE is set than it will be used over a CLI argument passed to the service.

## Ready-built images

Container images are published here:

https://ghcr.io/zapier/prom-aggregation-gateway

## Helm Chart

Helm Charts are published here:

https://zapier.github.io/prom-aggregation-gateway/

You can use them:

```
helm repo add pag https://zapier.github.io/prom-aggregation-gateway/
helm repo update
helm search repo pag -l
```

## Contributing
To run the server you can run:

```
go run .
```

### Testing
To run the tests you can run:

```
go test
```

### VSCode
To debug locally you can setup a launch.json with the following:

```
{
    "version": "0.2.0",
    "configurations": [
      {
        "name": "Debug prom-agg-gateway",
        "type": "go",
        "request": "launch",
        "mode": "debug",
        "program": "${workspaceFolder}",
        "args": [
          "run",
          "."
        ]
      }
    ]
  }
```

Then you'll be able to launch prom-agg-gateway locally and debug it from within VSCode.  You
must have [Delve](https://github.com/derekparker/delve) installed locally for this to work.

If you want a debugger that will launch the tests you can create this configuration:

```
{
    "version": "0.2.0",
    "configurations": [
      {
        "name": "Test Current File",
        "type": "go",
        "request": "launch",
        "mode": "test",
        "program": "${workspaceFolder}/${relativeFileDirname}",
        "showLog": true
      }
    ]
  }
  ```

## Comparison to [Prometheus Pushgateway](https://github.com/prometheus/pushgateway)

According to https://prometheus.io/docs/practices/pushing/:

> The Pushgateway never forgets series pushed to it and will expose them to Prometheus forever...
>
> The latter point is especially relevant when multiple instances of a job differentiate their metrics in the Pushgateway via an instance label or similar.

This restriction makes the Prometheus pushgateway inappropriate for the use case of accepting metrics from a client-side web app, so we created this one to aggregate counters from multiple senders.

Prom-aggregation-gateway presents a similar API but does not attempt to be a drop-in replacement.

## Client Libraries
### Python
- https://github.com/prometheus/client_python

### JS
- https://github.com/siimon/prom-client
- https://github.com/weaveworks/promjs/

## <a name="help"></a>Getting Help

If you have any questions about, feedback for or problems with `prom-aggregation-gateway`:

- [File an issue](https://github.com/zapier/prom-aggregation-gateway/issues/new).

prom-aggregation-gateway follows the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md). Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting a project maintainer.

Your feedback is always welcome!
