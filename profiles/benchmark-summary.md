Медианы пяти запусков; `-cpu=1 -p=1 -benchtime=200ms`. Полные измерения: `base-bench.txt`, `result-bench.txt`.

| Benchmark | ns/op до | ns/op после | B/op до | B/op после | allocs/op до | allocs/op после |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `BenchmarkBuildBatch` | 2277 | 1882 | 5384 | 5384 | 66 | 3 |
| `BenchmarkStorageSnapshot` | 1772 | 1763 | 3544 | 3544 | 4 | 4 |
| `BenchmarkCollector` | 14147 | 14517 | 984 | 984 | 4 | 4 |
| `BenchmarkCompress/1024` | 37727 | 26825 | 814096 | 240 | 20 | 3 |
| `BenchmarkCompress/4096` | 41260 | 29454 | 814096 | 240 | 20 | 3 |
| `BenchmarkCompress/65536` | 128880 | 125658 | 814736 | 880 | 21 | 4 |
| `BenchmarkDecompress` | 5039 | 5489 | 52368 | 52368 | 15 | 15 |
| `BenchmarkCompute` | 1696 | 1691 | 624 | 624 | 9 | 9 |
| `BenchmarkMemStorage/UpdateGauge` | 10.49 | 10.51 | 0 | 0 | 0 | 0 |
| `BenchmarkMemStorage/UpdateCounter` | 10.44 | 10.46 | 0 | 0 | 0 | 0 |
| `BenchmarkMemStorage/GetGauge` | 8.085 | 8.088 | 0 | 0 | 0 | 0 |
| `BenchmarkMemStorage/UpdateMetrics64` | 619.2 | 629.6 | 0 | 0 | 0 | 0 |
| `BenchmarkMemStorage/Snapshot64` | 1796 | 2066 | 3544 | 3544 | 4 | 4 |
| `BenchmarkMetricsPipeline` | 143674 | 128731 | 902668 | 88605 | 386 | 303 |
| `BenchmarkUpdatesHandler` | 2870 | 3877 | 7048 | 7048 | 39 | 39 |
| `BenchmarkPublisher/0` | 2.676 | 2.778 | 0 | 0 | 0 | 0 |
| `BenchmarkPublisher/1` | 3.784 | 3.863 | 0 | 0 | 0 | 0 |
| `BenchmarkPublisher/2` | 5.659 | 5.754 | 0 | 0 | 0 | 0 |
| `BenchmarkPublisher/8` | 15.21 | 15.25 | 0 | 0 | 0 | 0 |
| `BenchmarkFileObserver` | 1370 | 1390 | 48 | 48 | 1 | 1 |
