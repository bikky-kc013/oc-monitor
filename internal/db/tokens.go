package db

import (
	"database/sql"
	"strings"
)

func (d *DB) BatchInputTokens(sessionIDs []string) map[string]int64 {
	out := make(map[string]int64, len(sessionIDs))
	if len(sessionIDs) == 0 {
		return out
	}
	placeholders := strings.Repeat("?,", len(sessionIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(sessionIDs))
	for i, id := range sessionIDs {
		args[i] = id
	}
	rows, err := d.conn.Query(`
		SELECT m.session_id, json_extract(m.data,'$.tokens.input')
		FROM message m
		INNER JOIN (
			SELECT session_id, MAX(time_created) AS max_ts
			FROM message
			WHERE session_id IN (`+placeholders+`) AND json_extract(data,'$.role') = 'assistant'
			GROUP BY session_id
		) latest ON m.session_id = latest.session_id AND m.time_created = latest.max_ts`, args...)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var n sql.NullInt64
		if err := rows.Scan(&id, &n); err != nil {
			continue
		}
		if n.Valid {
			out[id] = n.Int64
		}
	}
	return out
}

func (d *DB) LatestInputTokens(sessionID string) int64 {
	var n sql.NullInt64
	d.conn.QueryRow(`
		SELECT json_extract(data,'$.tokens.input')
		FROM message WHERE session_id = ? AND json_extract(data,'$.role') = 'assistant'
		ORDER BY time_created DESC LIMIT 1`, sessionID).Scan(&n)
	if n.Valid {
		return n.Int64
	}
	return 0
}
