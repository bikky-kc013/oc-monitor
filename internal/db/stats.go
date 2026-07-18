package db

import (
	"database/sql"
	"path/filepath"
	"time"
)

func (d *DB) AgentStats() ([]AgentRow, error) {
	rows, err := d.conn.Query(`
		SELECT
			COALESCE(json_extract(data,'$.agent'), '')     AS agent,
			COALESCE(json_extract(data,'$.providerID'), '') AS provider_id,
			COALESCE(json_extract(data,'$.modelID'), '')    AS model_id,
			COUNT(DISTINCT session_id)                      AS sessions,
			SUM(COALESCE(json_extract(data,'$.cost'), 0))   AS cost,
			SUM(COALESCE(json_extract(data,'$.tokens.input'), 0)
			  + COALESCE(json_extract(data,'$.tokens.output'), 0)) AS tokens,
			MAX(time_created)                               AS last_active
		FROM message
		WHERE json_extract(data,'$.role') = 'assistant'
		GROUP BY agent, provider_id, model_id
		ORDER BY cost DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AgentRow
	for rows.Next() {
		var a AgentRow
		var providerID, modelID string
		var ts int64
		if err := rows.Scan(&a.Agent, &providerID, &modelID, &a.Count, &a.Cost, &a.Tokens, &ts); err != nil {
			return nil, err
		}
		a.Model = joinModel(providerID, modelID)
		a.LastActive = time.UnixMilli(ts)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (d *DB) ProjectRollups() ([]ProjectRow, error) {
	rows, err := d.conn.Query(sessionAgg + `
		SELECT
			s.directory,
			COUNT(*)                                                          AS sessions,
			SUM(COALESCE(agg.cost, 0))                                        AS cost,
			SUM(COALESCE(agg.tokens_input,0)+COALESCE(agg.tokens_output,0))   AS tokens,
			MAX(s.time_updated)                                               AS last_active
		FROM session s
		LEFT JOIN agg ON agg.session_id = s.id
		WHERE s.parent_id IS NULL OR s.parent_id = ''
		GROUP BY s.directory
		ORDER BY cost DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProjectRow
	for rows.Next() {
		var r ProjectRow
		var ts int64
		if err := rows.Scan(&r.Dir, &r.Count, &r.Cost, &r.Tokens, &ts); err != nil {
			return nil, err
		}
		r.FullPath = r.Dir
		r.Dir = filepath.Base(r.Dir)
		r.LastActive = time.UnixMilli(ts)

		var providerID, modelID string
		var discard int
		err := d.conn.QueryRow(`
			SELECT COALESCE(json_extract(m.data,'$.agent'), ''),
				COALESCE(json_extract(m.data,'$.providerID'), ''),
				COALESCE(json_extract(m.data,'$.modelID'), ''),
				COUNT(*) AS c
			FROM message m
			JOIN session s ON s.id = m.session_id
			WHERE s.directory = ? AND json_extract(m.data,'$.role') = 'assistant'
			GROUP BY 1, 2, 3 ORDER BY c DESC LIMIT 1`,
			r.FullPath).Scan(&r.Agent, &providerID, &modelID, &discard)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		r.Model = joinModel(providerID, modelID)
		out = append(out, r)
	}
	return out, rows.Err()
}
