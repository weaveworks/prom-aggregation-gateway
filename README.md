# Prometheus Aggregation Gateway

Prometheus Aggregation Gateway is a aggregating push gateway for Prometheus.  As opposed to the official [Prometheus Pushgateway](https://github.com/prometheus/pushgateway), this service aggregates the sample values it receives.

## Motivation

According to https://prometheus.io/docs/practices/pushing/:

> The Pushgateway never forgets series pushed to it and will expose them to Prometheus forever...
>
> The latter point is especially relevant when multiple instances of a job differentiate their metrics in the Pushgateway via an instance label or similar.

This restriction makes the pushgateway inappropriate for the usecase of accepting metrics from a client-side web app.

## JS Client Library

See https://github.com/weaveworks/promjs/ for a JS client library for Prometheus that can be used from within a web app.

## <a name="help"></a>Getting Help

If you have any questions about, feedback for or problems with `prom-aggregation-gateway`:

- Invite yourself to the <a href="https://weaveworks.github.io/community-slack/" target="_blank"> #weave-community </a> slack channel.
- Ask a question on the <a href="https://weave-community.slack.com/messages/general/"> #weave-community</a> slack channel.
- Send an email to <a href="mailto:weave-users@weave.works">weave-users@weave.works</a>
- <a href="https://github.com/weaveworks/prom-aggregation-gateway/issues/new">File an issue.</a>

Your feedback is always welcome!
