package db

func (d *DB) Search(term string, limit int) ([]SearchResult, error) {
	like := "%" + term + "%"
	rows, err := d.conn.Query(`
		SELECT p.session_id, s.title, json_extract(p.data,'$.type'),
			COALESCE(json_extract(p.data,'$.text'),'')
		FROM part p JOIN session s ON s.id = p.session_id
		WHERE json_extract(p.data,'$.type') IN ('text','reasoning')
			AND json_extract(p.data,'$.text') LIKE ?
		ORDER BY p.time_created DESC LIMIT ?`, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.SessionID, &r.SessionTitle, &r.PartType, &r.Text); err != nil {
			return nil, err
		}
		if len(r.Text) > 120 {
			r.Text = r.Text[:120] + "..."
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
