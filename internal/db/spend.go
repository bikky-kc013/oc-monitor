package db

import (
	"database/sql"
	"time"
)

func (d *DB) DailySpend(days int) ([]PeriodRow, error) {
	since := time.Now().AddDate(0, 0, -days).UnixMilli()
	rows, err := d.conn.Query(`
		SELECT
			date(time_created/1000, 'unixepoch') AS day,
			SUM(COALESCE(json_extract(data,'$.cost'), 0)) AS cost,
			SUM(COALESCE(json_extract(data,'$.tokens.input'),0)
			  + COALESCE(json_extract(data,'$.tokens.output'),0)) AS tokens
		FROM message
		WHERE json_extract(data,'$.role') = 'assistant' AND time_created >= ?
		GROUP BY 1 ORDER BY 1`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPeriods(rows)
}

func (d *DB) HourlySpend(hours int) ([]PeriodRow, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour).UnixMilli()
	rows, err := d.conn.Query(`
		SELECT
			strftime('%Y-%m-%d %H:00', time_created/1000, 'unixepoch') AS hour,
			SUM(COALESCE(json_extract(data,'$.cost'), 0)) AS cost,
			SUM(COALESCE(json_extract(data,'$.tokens.input'),0)
			  + COALESCE(json_extract(data,'$.tokens.output'),0)) AS tokens
		FROM message
		WHERE json_extract(data,'$.role') = 'assistant' AND time_created >= ?
		GROUP BY 1 ORDER BY 1`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPeriods(rows)
}

func (d *DB) TotalSpendSince(when time.Time) float64 {
	var n sql.NullFloat64
	d.conn.QueryRow(`
		SELECT SUM(COALESCE(json_extract(data,'$.cost'), 0))
		FROM message
		WHERE json_extract(data,'$.role') = 'assistant' AND time_created >= ?`,
		when.UnixMilli()).Scan(&n)
	if n.Valid {
		return n.Float64
	}
	return 0
}

func scanPeriods(rows *sql.Rows) ([]PeriodRow, error) {
	var out []PeriodRow
	for rows.Next() {
		var p PeriodRow
		if err := rows.Scan(&p.Label, &p.Cost, &p.Tokens); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
