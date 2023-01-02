import http from 'k6/http';
import { check } from 'k6';


const jobs = [
  "job1",
  "job2",
  "job3",
  "job4",
]

const labels = [
  "foo",
  "bar",
  "baz",
]


const labelValues = [
  "qux",
  "fred",
  "thud",
]

function getRandInteger(min, max) {
  return Math.floor(Math.random() * (max - min) ) + min;
}

export default function () {
  const randomJob = jobs[Math.floor(Math.random() * jobs.length)];
  const randomLabel = labels[Math.floor(Math.random() * labels.length)];
  const randomValue = labelValues[Math.floor(Math.random() * labelValues.length)];
  const url = `${__ENV.PAG_HOST}/metrics/job/${randomJob}/${randomLabel}/${randomValue}`;
  const randomMetric1 = getRandInteger(1, 50)
  const randomMetric2 = getRandInteger(1000, 3000)
  const randomMetric3 = getRandInteger(3000, 4000)

  const payload = `
  # TYPE some_metric counter
  some_metric{label="val1"} ${randomMetric1}
  # TYPE another_metric gauge
  # HELP another_metric Just an example.
  another_metric ${randomMetric2}
  k6_http_requests_total{method="post",code="200"} ${randomMetric3}
  `

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);
  check(res, { 'status was 202': (r) => r.status < 300 });
}
