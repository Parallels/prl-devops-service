package database

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactDSN_Matrix(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		want  string
		badIn string // substring that must NOT survive, if relevant
	}{
		{
			name:  "url percent-encoded password",
			in:    testDSN("u", "p%40ss", "h:5432", "db?sslmode=require"),
			want:  testDSN("u", "***", "h:5432", "db?sslmode=require"),
			badIn: "p%40ss",
		},
		{
			name:  "url raw-at password",
			in:    testDSN("u", "p@ss", "h", "db"),
			want:  testDSN("u", "***", "h", "db"),
			badIn: ":p@ss@",
		},
		{
			name: "url without password keeps user and params",
			in:   "postgres://u@h:5432/db?sslmode=disable",
			want: "postgres://u@h:5432/db?sslmode=disable",
		},
		{
			name:  "keyword form password",
			in:    "host=h user=u password=" + fixtureSecret("kw-pw") + " port=5432",
			want:  "host=h user=u password=*** port=5432",
			badIn: fixtureSecret("kw-pw"),
		},
		{
			name:  "keyword case-insensitive quoted with spaces",
			in:    "host=h PASSWD=" + "'a b c'",
			want:  "host=h PASSWD='***'",
			badIn: "a b c",
		},
		{
			name:  "sslkey and sslpassword scrubbed",
			in:    "host=h sslkey=/etc/pg/key.pem sslpassword=" + fixtureSecret("kw-ssl"),
			want:  "host=h sslkey=*** sslpassword=***",
			badIn: fixtureSecret("kw-ssl"),
		},
		{
			name:  "unicode keyword password",
			in:    "host=h password=" + "парол" + "66",
			want:  "host=h password=***",
			badIn: "парол",
		},
		{
			name:  "malformed url fails closed",
			in:    "postgres://%zz%zz\\bad",
			want:  "[redacted-dsn]",
			badIn: "%zz",
		},
		{
			name: "unrecognised non-empty collapses",
			in:   "hunter2",
			want: "[redacted-dsn]",
		},
		{name: "empty stays empty", in: "", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RedactDSN(tc.in)
			assert.Equal(t, tc.want, got)
			if tc.badIn != "" {
				assert.False(t, strings.Contains(got, tc.badIn), "secret fragment survived redaction")
			}
		})
	}
}
