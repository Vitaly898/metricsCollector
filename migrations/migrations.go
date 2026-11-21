// Package migrations предоставляет встроенные SQL-миграции схемы хранилища метрик.
package migrations

import "embed"

// FS содержит SQL-файлы миграций, встроенные в исполняемый файл.
//
//go:embed *.sql
var FS embed.FS
