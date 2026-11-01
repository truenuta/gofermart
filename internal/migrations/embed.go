// Package migrations хранит SQL-файлы миграций схемы базы данных,
// встроенные в бинарник на этапе сборки.
package migrations

import "embed"

// FS предоставляет доступ к встроенным файлам миграций (пары
// NNNNNN_name.up.sql / NNNNNN_name.down.sql). Используется как источник
// миграций для golang-migrate через source/iofs.
//
//go:embed *.sql
var FS embed.FS
