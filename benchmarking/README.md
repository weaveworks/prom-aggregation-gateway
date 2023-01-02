# Benchmarking
We are using [k6](https://k6.io/) for our benchmarking. To run one of the performance
tests you can do:

```
k6 run -e PAG_HOST=http://localhost --vus 10 --duration 30s  benchmarking/perf1.js
```