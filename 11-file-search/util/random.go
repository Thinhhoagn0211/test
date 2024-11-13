// util/helpers.go
package util

import (
	"database/sql"
	"time"
)

// NullableString handles empty strings as sql.NullString
func NullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// NullableInt64 handles 0 as sql.NullInt64
func NullableInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

// NullableTime handles zero-time as sql.NullTime
func NullableTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}
