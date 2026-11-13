package database

import (
	"net/url"
	"regexp"
	"strings"
)

// redactedDSNPlaceholder is the fail-closed collapse for any unrecognised non-empty
// DSN (asserted by name in the tests of both engines).
const redactedDSNPlaceholder = "[redacted-dsn]"

// dsnKeywordSecret matches libpq keyword-form secret keys and their values
// (bare, single- or double-quoted — quoted values may contain spaces).
var dsnKeywordSecret = regexp.MustCompile(`(?i)(password|passwd|sslpassword|sslkey)(\s*=\s*)("[^"]*"|'[^']*'|[^\s]*)`)

// dsnKeywordForm full-matches a well-formed libpq keyword/value string:
// every whitespace- or semicolon-separated token must be key=value (quoted
// values may contain spaces). Merely CONTAINING an "=" is not recognition —
// T16 showed a secret riding in malformed keyword-ish junk (e.g.
// "s3cret host=x") would otherwise echo through the scrub path untouched.
var dsnKeywordForm = regexp.MustCompile(
	`(?i)^\s*[A-Za-z_][A-Za-z0-9_]*\s*=\s*(?:"[^"]*"|'[^']*'|[^\s;]*)` +
		`(?:[;\s]+[A-Za-z_][A-Za-z0-9_]*\s*=\s*(?:"[^"]*"|'[^']*'|[^\s;]*))*[;\s]*$`)

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
			return redactedDSNPlaceholder
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
		return redactedDSNPlaceholder
	}

	// Recognition is strict keyword-form (see dsnKeywordForm): malformed
	// keyword-ish strings are UNRECOGNISED and collapse fail-closed — they
	// may be secrets glued to fragments, and no token-level scrub could
	// prove otherwise (T16).
	if !dsnKeywordForm.MatchString(dsn) {
		return redactedDSNPlaceholder
	}

	return scrubKeywordSecrets(dsn)
}

// scrubKeywordSecrets replaces the values of secret keys (password/passwd/
// sslpassword/sslkey, bare or quoted) with ***. Extracted so driver-text
// scrubbing can reuse the exact same token rule WITHOUT inheriting RedactDSN's
// DSN-recognition/collapse policy — error prose is not a DSN.
func scrubKeywordSecrets(s string) string {
	return dsnKeywordSecret.ReplaceAllStringFunc(s, func(match string) string {
		parts := dsnKeywordSecret.FindStringSubmatch(match)
		quote := ""
		if v := parts[3]; strings.HasPrefix(v, "'") || strings.HasPrefix(v, `"`) {
			quote = v[:1]
		}
		return parts[1] + parts[2] + quote + "***" + quote
	})
}

// scrubDriverError removes the raw DSN from driver-produced text before it
// reaches a log line (T16). pgx echoes unparseable inputs verbatim inside
// `cannot parse ...` messages, and its own masking only covers the
// keyword/password token shape. Two passes, both shape-safe for prose:
//  1. exact-input replacement — kills the verbatim echo; the substitute is
//     RedactDSN(rawDSN), which is itself fail-closed (placeholder for junk,
//     token-scrubbed view for well-formed DSNs);
//  2. token-level scrubKeywordSecrets — catches pgx-normalized echo variants
//     that carry password=/sslkey= tokens; deliberately NOT RedactDSN, whose
//     recognition would collapse legitimate dial-failure prose (ops needs
//     the addresses and retry detail there).
func scrubDriverError(errText, rawDSN string) string {
	safe := errText
	if rawDSN != "" {
		safe = strings.ReplaceAll(safe, rawDSN, RedactDSN(rawDSN))
	}
	return scrubKeywordSecrets(safe)
}
