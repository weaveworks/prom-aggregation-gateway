# Prometheus Aggregation Gateway

Prometheus Aggregation Gateway is a aggregating push gateway for Prometheus.  As opposed to the official [Prometheus Pushgateway](https://github.com/prometheus/pushgateway), this service aggregates the sample values it receives.

* Counters where all labels match are added up.
* Histograms are added up; if bucket boundaries are mismatched then the result has the union of all buckets and counts are given to the lowest bucket that fits.
* Gauges are also added up (but this may not make any sense)
* Summaries are discarded.

## Ready-built images

Available on DockerHub `weaveworks/prom-aggregation-gateway`

## Motivation

According to https://prometheus.io/docs/practices/pushing/:

> The Pushgateway never forgets series pushed to it and will expose them to Prometheus forever...
>
> The latter point is especially relevant when multiple instances of a job differentiate their metrics in the Pushgateway via an instance label or similar.

This restriction makes the Prometheus pushgateway inappropriate for the usecase of accepting metrics from a client-side web app, so we created this one to aggregate counters from multiple senders.

## JS Client Library

See https://github.com/weaveworks/promjs/ for a JS client library for Prometheus that can be used from within a web app.

## <a name="help"></a>Getting Help

If you have any questions about, feedback for or problems with `prom-aggregation-gateway`:

- Invite yourself to the <a href="https://slack.weave.works/" target="_blank">Weave Users Slack</a>.
- Ask a question on the [#general](https://weave-community.slack.com/messages/general/) slack channel.
- [File an issue](https://github.com/weaveworks/prom-aggregation-gateway/issues/new).

Your feedback is always welcome!
