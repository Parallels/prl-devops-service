package database

import (
	"net/url"
	"regexp"
	"strings"
)

// dsnKeywordSecret matches libpq keyword-form secret keys and their values
// (bare, single- or double-quoted — quoted values may contain spaces).
var dsnKeywordSecret = regexp.MustCompile(`(?i)(password|passwd|sslpassword|sslkey)(\s*=\s*)("[^"]*"|'[^']*'|[^\s]*)`)

// RedactDSN removes credentials from a Postgres connection string for safe
// logging. Handles libpq keyword form and URL form; anything unrecognised and
// non-empty collapses to "[redacted-dsn]" — fail-closed, never echo input.
//
// Deliberate choice (reviewed): URL usernames are KEPT — they are needed for
// ops triage and are not secrets in this system's threat model; passwords,
// sslkey paths and sslpasswords never survive.
func RedactDSN(dsn string) string {
	if dsn == "" {
		return ""
	}

	if strings.Contains(dsn, "://") {
		u, err := url.Parse(dsn)
		if err != nil || u.Host == "" {
			return "[redacted-dsn]"
		}
		if _, hasPass := u.User.Password(); !hasPass {
			return dsn // URL form without credentials: nothing to scrub
		}
		// Splice manually rather than rebuild via url.String():
		// url.UserPassword would percent-escape "***" to %2A%2A%2A, and manual
		// splicing keeps everything outside the userinfo byte-identical.
		// Go's url.Parse splits userinfo at the LAST '@' (libpq convention),
		// so raw-@ passwords are correctly located before we splice.
		i := strings.Index(dsn, "://")
		after := dsn[i+3:]
		end := len(after)
		if s := strings.IndexAny(after, "/?#"); s >= 0 {
			end = s
		}
		auth := after[:end]
		at := strings.LastIndex(auth, "@")
		user := auth[:at]
		if colon := strings.Index(user, ":"); colon >= 0 {
			user = user[:colon] // keep username (ops triage), drop password bytes
		}
		return dsn[:i+3] + user + ":***@" + after[at+1:]
	}

	if !strings.Contains(dsn, "=") {
		return "[redacted-dsn]"
	}

	return dsnKeywordSecret.ReplaceAllStringFunc(dsn, func(match string) string {
		parts := dsnKeywordSecret.FindStringSubmatch(match)
		quote := ""
		if v := parts[3]; strings.HasPrefix(v, "'") || strings.HasPrefix(v, `"`) {
			quote = v[:1]
		}
		return parts[1] + parts[2] + quote + "***" + quote
	})
}
