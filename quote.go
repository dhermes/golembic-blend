package golembic

import (
	"strings"
)

// QuoteIdentifier quotes an identifier, such as a table name, for usage
// in a query.
//
// This implementation is vendored in here to avoid the side effects of
// importing `github.com/lib/pq`.
//
// See:
// - https://github.com/lib/pq/blob/v1.8.0/conn.go#L1564-L1581
// - https://www.sqlite.org/lang_keywords.html
// - https://github.com/ronsavage/SQL/blob/a67e7eaefae89ed761fa4dcbc5431ec9a235a6c8/sql-99.bnf#L412
func QuoteIdentifier(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return `"` + strings.Replace(name, `"`, `""`, -1) + `"`
}
