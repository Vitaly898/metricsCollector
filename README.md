# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Бенчмарки и профили памяти — iter17

Бенчмарки находятся в `internal/{agent,audit,compress,hash,server,storage}/benchmark_test.go`.
Они измеряют сборку пакета и снимка метрик агента, сбор runtime-метрик, gzip,
подпись HMAC, операции хранилища, обработку HTTP-запросов и аудит.

Запуск из корня проекта:

```sh
go test -p=1 -run '^$' -bench . -benchmem -benchtime=200ms -count=5 -cpu=1 \
  ./internal/agent ./internal/compress ./internal/hash \
  ./internal/storage ./internal/server ./internal/audit
```

`ns/op` — время операции, `B/op` — выделенные байты, `allocs/op` — число аллокаций.
`-p=1` запускает пакеты последовательно, `-cpu=1` задаёт один поток выполнения Go;
измерения проводятся без `-race`. Условия: Go 1.26.0, darwin/arm64, Apple M3 Pro.
Ниже приведены медианы пяти запусков. Время зависит от нагрузки машины;
изменения времени и аллокаций рассматриваются отдельно.

| Компонент | ns/op до → после | B/op до → после | allocs/op до → после |
| --- | ---: | ---: | ---: |
| Сборка пакета из 64 gauge и одного counter | 2277 → 1882 | 5384 → 5384 | 66 → 3 |
| Сжатие 4096 байт | 41260 → 29454 | 814096 → 240 | 20 → 3 |
| Полный путь пакета из 65 метрик | 143674 → 128731 | 902668 → 88605 | 386 → 303 |

Полные исходные замеры: [до](profiles/base-bench.txt), [после](profiles/result-bench.txt).
[Таблица всех бенчмарков](profiles/benchmark-summary.md) включает и неизменённые компоненты.

### Как снимались профили

`TestMemoryProfile` в `internal/server/benchmark_test.go` выполняет **ровно 1000**
пакетов по 65 метрик. Каждый пакет проходит через настоящие
`Sender.BuildBatch`, JSON, gzip, HMAC, маршрутизатор сервера, middleware и
`MemStorage`. Проверяются HTTP-статус и итоговое значение счётчика.
Сеть, PostgreSQL, файловое хранилище и приёмники аудита в этот сценарий не входят;
аудит отдельно измеряется бенчмарками. Запрос и ответ создаются через `httptest`.
Это измерение выбранного пути выполнения, а не RSS работающего процесса.

Оба профиля получены при одинаковой нагрузке, `GOMAXPROCS=1` и
`-memprofilerate=1`; после обработки вызывается `runtime.GC()`.
Используется `pprof.Lookup("allocs")`: профиль содержит `alloc_space`,
`alloc_objects`, `inuse_space`, `inuse_objects`, а по умолчанию показывает
`alloc_space` — все выделенные за запуск байты. Поэтому команда сравнения ниже
оценивает сокращение аллокаций. Обычные тесты пропускают сбор профиля.

```sh
mkdir -p profiles
METRICS_PROFILE="$PWD/profiles/result.pprof" \
  go test -run '^TestMemoryProfile$' -count=1 -memprofilerate=1 ./internal/server
```

`profiles/base.pprof` снят **до** правок производственного кода,
`profiles/result.pprof` — после. Чтобы повторить исходную нагрузку на версии
`iter16`, можно создать временную копию коммита `293ffd6` и добавить в неё
те же измерительные файлы, сохранив текущую ветку:

```sh
project_dir="$PWD"
baseline_dir="$(mktemp -d)"
git archive 293ffd6 | tar -x -C "$baseline_dir"
for package in agent audit compress hash server storage; do
  cp "internal/$package/benchmark_test.go" "$baseline_dir/internal/$package/"
done
(
  cd "$baseline_dir"
  METRICS_PROFILE="$project_dir/profiles/base.pprof" \
    go test -run '^TestMemoryProfile$' -count=1 -memprofilerate=1 ./internal/server
  go test -p=1 -run '^$' -bench . -benchmem -benchtime=200ms -count=5 -cpu=1 \
    ./internal/agent ./internal/compress ./internal/hash \
    ./internal/storage ./internal/server ./internal/audit \
    > "$project_dir/profiles/base-bench.txt"
)
```

Эти команды перезаписывают артефакты `base`. Для итоговых замеров используется
текущая ветка и те же команды бенчмарков с выводом в `profiles/result-bench.txt`.

### Анализ pprof и выбранные изменения

В окружении не было готового `go tool pprof`, поэтому инструмент собран
из исходников установленной версии Go. Эквивалентный запуск:

```sh
go build -o /tmp/metrics-iter17-pprof cmd/pprof
pprof() { /tmp/metrics-iter17-pprof "$@"; }

pprof -top profiles/base.pprof
pprof -top -sample_index=alloc_objects profiles/base.pprof
pprof -top -sample_index=inuse_space profiles/base.pprof
pprof -list='metricsCollector/internal/compress.Compress' profiles/base.pprof
pprof -peek='compress/flate.NewWriter|compress/gzip.NewReader|Sender.*BuildBatch' profiles/base.pprof
pprof -web -output=profiles/base.svg profiles/base.pprof
pprof -web -output=profiles/result.svg profiles/result.pprof
```

Для `web` нужен Graphviz (`dot`; на macOS: `brew install graphviz`).
`-output` сохраняет граф в SVG без запуска браузера.
В режиме `pprof profiles/base.pprof` доступны и интерактивные команды
`top`, `list`, `peek`, `web`.

`top` выделил основные источники аллокаций, `list` связал их со строками
`Compress`, `peek` и граф `web` показали цепочку вызовов до `compress/flate.NewWriter`.
На gzip-сжатие приходилось около 90,21% выделенных байтов.
По числу объектов выделялся также `Sender.BuildBatch`.

Изменены два участка:

1. В `internal/compress/gzip.go` добавлен `sync.Pool` для gzip-компрессоров.
   Вместо создания рабочих структур deflate на каждой отправке компрессор
   переиспользуется через `Reset`. Перед возвратом в пул он переключается на
   `io.Discard`, чтобы не удерживать буфер результата. Каждый вызов получает
   собственный выходной буфер, поэтому следующие вызовы не меняют предыдущие
   результаты. Интерфейсы `NewWriter` и `NewReader` сохранены.
2. В `internal/agent/sender.go` значения gauge размещаются в одном массиве.
   Поля `Value` указывают на разные элементы этого массива. Пакет остаётся
   независимым от исходной map и последующих пакетов, но вместо отдельной
   аллокации на каждый gauge используется один массив. Объём полезных данных
   почти тот же — улучшилось число аллокаций.

Копии снимков хранилища сохраняются: они нужны для безопасной работы с данными
за пределами блокировки. Распаковка gzip остаётся заметной в итоговом профиле;
в этой итерации она не менялась.

### Сравнение памяти

| Показатель за одинаковую нагрузку | До | После |
| --- | ---: | ---: |
| `alloc_space`, выделено за весь запуск | 861,32 MiB | 85,76 MiB |
| `alloc_objects`, объекты в профиле | 255179 | 204632 |
| `inuse_space`, живые объекты после GC | 484,36 KiB | 1278,20 KiB |

Выделенные байты сократились на **90,04%**: с 903162776 до 89929200 байт.
Удерживаемая память выросла примерно на **794 KiB**: в этом последовательном
сценарии пул сохраняет один компрессор с рабочими структурами. При параллельном
сжатии компрессоров может быть больше; GC вправе очистить пул.
Это обмен небольшого сохраняемого рабочего набора на уменьшение постоянных
выделений памяти. Уменьшение `alloc_space` не означает уменьшение живого heap.

Команда из задания:

```sh
pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

Фактический вывод:

```text
File: server.test
Type: alloc_space
Time: 2026-11-07 19:16:12 MSK
Showing nodes accounting for -771.44MB, 89.57% of 861.32MB total
Dropped 72 nodes (cum <= 4.31MB)
      flat  flat%   sum%        cum   cum%
 -632.18MB 73.40% 73.40%  -768.64MB 89.24%  compress/flate.NewWriter (inline)
 -132.68MB 15.40% 88.80%  -132.68MB 15.40%  compress/flate.(*compressor).initDeflate (inline)
   -6.59MB  0.76% 89.57%    -6.59MB  0.76%  compress/flate.(*huffmanEncoder).generate
         0     0% 89.57%    -6.59MB  0.76%  compress/flate.(*Writer).Close (inline)
         0     0% 89.57%    -6.59MB  0.76%  compress/flate.(*compressor).close
         0     0% 89.57%    -6.59MB  0.76%  compress/flate.(*compressor).deflate
         0     0% 89.57%  -136.46MB 15.84%  compress/flate.(*compressor).init
         0     0% 89.57%    -6.59MB  0.76%  compress/flate.(*compressor).writeBlock
         0     0% 89.57%    -4.39MB  0.51%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 89.57%    -6.59MB  0.76%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0% 89.57%    -6.59MB  0.76%  compress/gzip.(*Writer).Close
         0     0% 89.57%  -768.64MB 89.24%  compress/gzip.(*Writer).Write
         0     0% 89.57%  -775.37MB 90.02%  github.com/Vitaly898/metricsCollector/internal/compress.Compress
         0     0% 89.57%  -775.55MB 90.04%  github.com/Vitaly898/metricsCollector/internal/server.(*metricsWorkload).run
         0     0% 89.57%  -775.55MB 90.04%  github.com/Vitaly898/metricsCollector/internal/server.TestMemoryProfile
         0     0% 89.57%  -775.55MB 90.04%  testing.tRunner
```

Отрицательные значения показывают уменьшение выделений на соответствующих
функциях. `flat` — собственный вклад функции, `cum` включает вызываемые функции;
строки `cum` нельзя складывать из-за пересекающихся стеков.

Сохранены [текст сравнения](profiles/diff.txt), отчёты
[исходных аллокаций](profiles/base-top.txt),
[исходных объектов](profiles/base-objects.txt),
[исходной живой памяти](profiles/base-inuse.txt),
[list](profiles/base-list.txt), [peek](profiles/base-peek.txt),
[итоговых аллокаций](profiles/result-top.txt),
[итоговых объектов](profiles/result-objects.txt),
[итоговой живой памяти](profiles/result-inuse.txt)
и графы [до](profiles/base.svg), [после](profiles/result.svg).

### Проверки корректности

Выполнены:

```sh
go test -race ./...
go vet ./...
go build ./cmd/agent ./cmd/server
```

Новые тесты проверяют round trip gzip, пустые и повреждённые данные,
независимость результатов при переиспользовании компрессора, параллельные
вызовы и независимость значений разных пакетов агента. Существующие проверки
передачи gzip, подписей, повторных отправок и аудита также прошли.
Дополнительно выполнен официальный `TestIteration14` из `metricstest_v2 v0.13.8`
с отдельной PostgreSQL 16. Проверка прошла после исправления обработки запросов
без необязательного заголовка `HashSHA256`. Регрессионные тесты проверяют JSON,
gzip и подпись ответов при включённом ключе. Полный внешний набор автотестов
не запускался; PostgreSQL в измерения памяти не включён.

## Godoc и примеры эндпоинтов — iter19

Экспортированные типы, интерфейсы, методы, функции, поля и константы описаны
godoc-комментариями непосредственно в Go-файлах. Просмотр документации:

```sh
go doc ./internal/handler
go doc ./internal/handler.UpdateJSONHandler.ServeHTTP
go doc ./internal/storage.MemStorage
```

В [internal/handler/example_test.go](internal/handler/example_test.go) находятся
исполняемые примеры с `// Output:`, оформленные по
[образцу стандартной библиотеки Go](https://go.dev/src/sort/example_interface_test.go).
Внешний пакет `handler_test` использует публичный API обработчиков, `httptest`
и хранилище в памяти; запуск сервера, PostgreSQL или сетевые соединения не требуются.

| Эндпоинт | Пример |
| --- | --- |
| `POST /update/{type}/{name}/{value}` | `ExampleUpdateHandler_ServeHTTP` |
| `POST /update/` | `ExampleUpdateJSONHandler_ServeHTTP` |
| `POST /updates/` | `ExampleUpdatesHandler_ServeHTTP` |
| `GET /value/{type}/{name}` | `ExampleValueHandler_ServeHTTP` |
| `POST /value/` | `ExampleValueJSONHandler_ServeHTTP` |
| `GET /ping` | `ExamplePingHandler_ServeHTTP` |
| `GET /` | `ExampleIndexHandler_ServeHTTP` |

Запуск всех семи примеров с проверкой фактического вывода:

```sh
go test -v ./internal/handler -run '^Example' -count=1
```

Примеры также выполняются при обычном `go test ./...`.
