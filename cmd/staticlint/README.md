# staticlint

Общая точка запуска анализаторов через `golang.org/x/tools/go/analysis/multichecker`.

```sh
go run ./cmd/staticlint ./...
```

Состав: стандартный набор `go vet` для Go 1.26, все проверки Staticcheck классов
SA, S, ST и QF, `ineffassign`, `bodyclose` и собственный `exitcheck`.

GoDoc содержит запуск и назначение каждой включённой проверки:

```sh
go doc ./cmd/staticlint
go run ./cmd/staticlint help SA1000
```

Для интеграции с `go vet`:

```sh
go build -o /tmp/staticlint ./cmd/staticlint
go vet -vettool=/tmp/staticlint ./...
```

`exitcheck` запрещает встроенную `panic` везде, `os.Exit` в `main.main` и
`log.Fatal`, `log.Fatalf`, `log.Fatalln` вне `main.main`. Вложенные анонимные
функции проверяются как самостоятельные функции. Для отдельного запуска только
этих правил остаётся команда `cmd/linter`.

```sh
go test ./cmd/staticlint ./internal/analyzer/exitcheck
```
