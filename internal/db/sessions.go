package db

import (
	"database/sql"
	"fmt"
	"time"
)

const sessionAgg = `
WITH asst AS (
  SELECT
    session_id,
    time_created,
    json_extract(data,'$.agent')      AS agent,
    json_extract(data,'$.providerID') AS provider_id,
    json_extract(data,'$.modelID')    AS model_id,
    COALESCE(json_extract(data,'$.cost'), 0)             AS cost,
    COALESCE(json_extract(data,'$.tokens.input'), 0)     AS tok_in,
    COALESCE(json_extract(data,'$.tokens.output'), 0)    AS tok_out,
    COALESCE(json_extract(data,'$.tokens.reasoning'), 0) AS tok_reason
  FROM message
  WHERE json_extract(data,'$.role') = 'assistant'
),
agg AS (
  SELECT
    session_id,
    SUM(cost)         AS cost,
    SUM(tok_in)        AS tokens_input,
    SUM(tok_out)        AS tokens_output,
    SUM(tok_reason)     AS tokens_reasoning,
    MAX(time_created)   AS last_msg_time
  FROM asst
  GROUP BY session_id
),
latest AS (
  SELECT a.session_id, a.agent, a.provider_id, a.model_id
  FROM asst a
  JOIN agg g ON g.session_id = a.session_id AND g.last_msg_time = a.time_created
)
`

const sessionSelectCols = `
	s.id, s.title, s.directory,
	COALESCE(latest.agent, ''),
	COALESCE(latest.provider_id, ''), COALESCE(latest.model_id, ''),
	COALESCE(agg.cost, 0), COALESCE(agg.tokens_input, 0),
	COALESCE(agg.tokens_output, 0), COALESCE(agg.tokens_reasoning, 0),
	s.time_created, s.time_updated, s.time_compacting
`

const sessionJoins = `
	FROM session s
	LEFT JOIN agg ON agg.session_id = s.id
	LEFT JOIN latest ON latest.session_id = s.id
`

func scanSessions(rows *sql.Rows) ([]Session, error) {
	var out []Session
	for rows.Next() {
		var s Session
		var providerID, modelID string
		var created, updated int64
		var compacting sql.NullInt64
		if err := rows.Scan(&s.ID, &s.Title, &s.Directory, &s.Agent,
			&providerID, &modelID,
			&s.Cost, &s.TokensIn, &s.TokensOut, &s.TokensReason,
			&created, &updated, &compacting); err != nil {
			return nil, err
		}
		s.ModelID = joinModel(providerID, modelID)
		s.Created = time.UnixMilli(created)
		s.Updated = time.UnixMilli(updated)
		if compacting.Valid {
			t := time.UnixMilli(compacting.Int64)
			s.Compacting = &t
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (d *DB) Sessions(limit int) ([]Session, error) {
	rows, err := d.conn.Query(sessionAgg+`SELECT `+sessionSelectCols+sessionJoins+`
		ORDER BY s.time_updated DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

func (d *DB) ActiveSessions(threshold time.Duration) ([]Session, error) {
	cutoff := time.Now().Add(-threshold).UnixMilli()
	rows, err := d.conn.Query(sessionAgg+`SELECT `+sessionSelectCols+sessionJoins+`
		WHERE s.time_updated > ? ORDER BY s.time_updated DESC`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

func (d *DB) SessionsByDir(dir string, limit int) ([]Session, error) {
	rows, err := d.conn.Query(sessionAgg+`SELECT `+sessionSelectCols+sessionJoins+`
		WHERE s.directory = ? ORDER BY s.time_updated DESC LIMIT ?`, dir, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

func (d *DB) GetSession(id string) (*Session, error) {
	rows, err := d.conn.Query(sessionAgg+`SELECT `+sessionSelectCols+sessionJoins+`
		WHERE s.id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ss, err := scanSessions(rows)
	if err != nil {
		return nil, err
	}
	if len(ss) == 0 {
		return nil, fmt.Errorf("not found")
	}
	return &ss[0], nil
}
