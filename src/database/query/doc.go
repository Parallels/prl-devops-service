// Package query is the transcription of the legacy Mongo-flavoured query language
// (data/filter_layer.go, data/order_layer.go, data/index_layer.go) for use by the
// GORM-backed engines of PBI #421.
//
// Parity strategy (#421 plan D-4, risk R-3 mitigation, taken literally):
// these functions are a transcription, NOT a rewrite — the bodies mirror the
// legacy layer verbatim, quirks preserved on purpose:
//   - filter values are unanchored Go regexps, not equality;
//   - the "i" option wraps the pattern in a "(?i)" prefix (repeated options
//     double-prefix — legal, deliberately not "fixed"; the prefix also
//     accumulates across objects within one call, exactly like legacy);
//   - non-string scalars compare through fmt.Sprintf("%v", …);
//   - struct-kind properties never match; GetRecordIndex bool/numeric
//     branches are legacy dead code and stay dead;
//   - OrderByProperty sorts the caller's slice in place and non-string
//     fields order as reflect String()'s non-string placeholder.
//
// NO-SQL-PUSH-DOWN RULE (D-4): domain code evaluates these functions in Go
// over GORM-loaded slices. SQL WHERE/ORDER BY push-down is opt-in per shape
// and only ever allowed for a shape that carries a dedicated differential
// test in the same PR. Never push regexp semantics into SQL: SQLite has no
// REGEXP without a UDF and Postgres "~" is not RE2.
//
// Intentional deltas from legacy (both pinned by golden tests):
//  1. getProperty starts with reflect.Indirect + an IsValid guard, so pointer
//     element types (*T) and nil pointers resolve instead of panicking on
//     NumField — behaviour for value types is unchanged (Indirect is a no-op
//     on struct values).
package query
