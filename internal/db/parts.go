package db

import (
	"strings"
	"time"
)

func (d *DB) Parts(sessionID string, limit int) ([]Part, error) {
	rows, err := d.conn.Query(`
		SELECT id, json_extract(data,'$.type'), COALESCE(json_extract(data,'$.tool'),''),
			COALESCE(json_extract(data,'$.text'),''),
			CASE WHEN json_extract(data,'$.state.output') IS NOT NULL THEN 1 ELSE 0 END,
			time_created
		FROM part WHERE session_id = ? ORDER BY time_created DESC LIMIT ?`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Part
	for rows.Next() {
		var p Part
		var ho int
		var ts int64
		if err := rows.Scan(&p.ID, &p.Type, &p.Tool, &p.Text, &ho, &ts); err != nil {
			return nil, err
		}
		p.HasOutput = ho == 1
		p.Created = time.UnixMilli(ts)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d *DB) LatestActivity(sessionIDs []string) (map[string]Activity, error) {
	if len(sessionIDs) == 0 {
		return map[string]Activity{}, nil
	}
	placeholders := strings.Repeat("?,", len(sessionIDs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(sessionIDs))
	for i, id := range sessionIDs {
		args[i] = id
	}
	rows, err := d.conn.Query(`
		SELECT p.session_id,
			json_extract(p.data,'$.type'),
			COALESCE(json_extract(p.data,'$.tool'),''),
			CASE WHEN json_extract(p.data,'$.state.output') IS NOT NULL THEN 1 ELSE 0 END,
			p.time_created
		FROM part p
		INNER JOIN (
			SELECT session_id, MAX(time_created) AS max_ts
			FROM part WHERE session_id IN (`+placeholders+`)
			GROUP BY session_id
		) latest ON p.session_id = latest.session_id AND p.time_created = latest.max_ts`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]Activity, len(sessionIDs))
	for rows.Next() {
		var a Activity
		var ho int
		var ts int64
		if err := rows.Scan(&a.SessionID, &a.Type, &a.Tool, &ho, &ts); err != nil {
			return nil, err
		}
		a.HasOutput = ho == 1
		a.Created = time.UnixMilli(ts)
		out[a.SessionID] = a
	}
	return out, rows.Err()
}
