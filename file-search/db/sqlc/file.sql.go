// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: file.sql

package db

import (
	"context"
	"database/sql"
	"time"
)

const getFile = `-- name: GetFile :many
SELECT 
    attributes
FROM 
    file
WHERE 
    -- File name (fuzzy search with index on name column)
    (name ILIKE '%' || $1 || '%' OR $1 IS NULL)
    
    -- Exact match for file extension (index on extension column)
    AND (extension = $2 OR $2 IS NULL)
    
    -- File size range (index on size column)
    AND (size >= $3 OR $3 IS NULL)
    AND (size <= $4 OR $4 IS NULL)
    
    -- -- Creation date range (index on created_at column)
    AND (created_at >= COALESCE($5, '1970-01-01 00:00:00'::timestamp))
    AND (created_at <= COALESCE($6, CURRENT_TIMESTAMP))
`

type GetFileParams struct {
	Column1     sql.NullString `json:"column_1"`
	Extension   string         `json:"extension"`
	Size        int64          `json:"size"`
	Size_2      int64          `json:"size_2"`
	CreatedAt   time.Time      `json:"created_at"`
	CreatedAt_2 time.Time      `json:"created_at_2"`
}

func (q *Queries) GetFile(ctx context.Context, arg GetFileParams) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, getFile,
		arg.Column1,
		arg.Extension,
		arg.Size,
		arg.Size_2,
		arg.CreatedAt,
		arg.CreatedAt_2,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []string{}
	for rows.Next() {
		var attributes string
		if err := rows.Scan(&attributes); err != nil {
			return nil, err
		}
		items = append(items, attributes)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
