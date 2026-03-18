package repository

import (
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

/*
Numeric -> string
*/
func numericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0"
	}

	f, _ := n.Float64Value()
	return strconv.FormatFloat(f.Float64, 'f', -1, 64)
}

/*
string -> pgtype.Text
*/
func stringToText(s string) pgtype.Text {
	pgText := pgtype.Text{
		String: s,
		Valid: true,
	}

	return pgText
}


/*
string -> Numeric
*/
func stringToNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(s)
	return n, err
}

/*
Timestamp -> time.Time
*/
func timestampToTime(t pgtype.Timestamp) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

/*
Timestamptz -> time.Time
*/
func timestamptzToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}