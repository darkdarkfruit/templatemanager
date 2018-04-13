## Benchmark
```bash
# debug-mode
> wrk -d5 'http://localhost:10000/demo1'
Running 5s test @ http://localhost:10000/demo1
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.97ms    5.06ms  50.28ms   86.82%
    Req/Sec     1.27k   295.83     2.61k    76.00%
  12695 requests in 5.01s, 53.21MB read
Requests/sec:   2532.09
Transfer/sec:     10.61MB

```

```bash
# production-mode
> wrk -d5 'http://localhost:10000/demo1'
Running 5s test @ http://localhost:10000/demo1
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.21ms    2.32ms  36.00ms   90.58%
    Req/Sec    10.09k     1.82k   15.99k    70.00%
  100514 requests in 5.01s, 421.08MB read
Requests/sec:  20073.50
Transfer/sec:     84.09MB

```