package database

import (
	"strings"

	"gorm.io/gorm"
)

// Insertion order, portably.
//
// The legacy store was an ordered array, and the service still hands lists to
// the API in the order rows were appended — catalog manifests, jobs, hosts,
// user configs, and the role/claim ordering that the #421 D-4 tests pin. SQLite
// gives that order for free as `rowid`. PostgreSQL has no rowid, and every query
// that asks for one fails outright ("column "rowid" does not exist"), which is
// how a fresh PostgreSQL install ends up unable to boot.
//
// What this returns on PostgreSQL is the table's primary key, which yields a
// *stable, deterministic* order — and, being explicit about it, an order the
// tests can pin. Read the following before assuming it is the same order:
//
// The ids here are random 32-hex strings (helpers.GenerateId), not time-ordered,
// so `ORDER BY id` is not insertion order. Three consequences, in rough order of
// how much they matter:
//
//   - List position changes relative to SQLite. Same rows, different order, no
//     error. Anything that sorted itself by rowid after reading is unaffected;
//     anything that showed the operator "what was added first" now shows
//     whatever the hex happens to sort to.
//   - A role or claim rename rewrites its id (UpdateClaim/UpdateRole keep the
//     row in place by updating the id column, which on SQLite leaves the rowid —
//     and so the position — untouched). On PostgreSQL the renamed row moves in
//     the list. Preserving that would need a real insertion-sequence column, not
//     a different ORDER BY expression; it is a schema change, deliberately not
//     smuggled in here.
//   - The singleton tables (configurations, reverse_proxies, schema_metadata)
//     hold one row by invariant, so the choice only decides which row wins when
//     that invariant is already broken — which is what those reads warn about.
//
// SQLite keeps `rowid` verbatim: the supported backend, its behaviour, and its
// existing tests are untouched by anything in this file.

// insertionOrder is the ORDER BY expression for "oldest first" on the engine
// attached to db. Callers pass the table they are ordering, because most sites
// query through Table("name") where GORM has no model to ask, and a Preload
// orders the *related* table rather than the one its parent query named.
func insertionOrder(db *gorm.DB, table string) string {
	// Through dialectOf, the one place that asks the driver what it is. An
	// unknown engine gets the supported backend's answer: SQLite.
	return insertionOrderFor(dialectOf(db), table)
}

// insertionOrderFor is the decision, split out so it can be tested against both
// engines without a server: the PostgreSQL leg of a live connection cannot be
// stood up in a unit test, but the string it produces can.
func insertionOrderFor(dialect, table string) string {
	if dialect != "postgres" {
		return "rowid"
	}

	if ordered, ok := postgresOrder[table]; ok {
		return ordered
	}

	// Qualified, because Preload builds a join: a bare "id" would be ambiguous
	// the moment two tables in the statement both have one.
	return table + ".id"
}

// postgresOrder lists the exceptions — tables whose primary key is not `id`.
// Everything else orders by `<table>.id`.
var postgresOrder = map[string]string{
	"schema_metadata": "schema_metadata.version", // keyed by version, not id
}

// preloadOrder orders a preload by the table it loads, not by the table the
// parent query named: Preload("Roles.Claims") runs its query against claims.
// The path's last segment is the relationship name, which is the entity's
// plural and therefore, for every model here, its table name.
func preloadOrder(db *gorm.DB, path string) string {
	return insertionOrder(db, relatedTable(path))
}

func relatedTable(path string) string {
	segments := strings.Split(path, ".")
	last := strings.ToLower(segments[len(segments)-1])

	if table, ok := relationshipTables[last]; ok {
		return table
	}

	return last
}

// relationshipTables covers a relationship whose name is not its table name.
// There are none today; the map exists so the day one appears, it is a data
// addition and not a silently wrong ORDER BY.
var relationshipTables = map[string]string{}
