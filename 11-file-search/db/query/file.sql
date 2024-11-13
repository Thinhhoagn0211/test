-- name: GetFiles :many
SELECT 
    path
FROM 
    file
WHERE 
    -- Filename fuzzy search (using LIKE for approximate matching)
    (name ILIKE '%' || $1 || '%' OR $1 IS NULL)
    
    -- File extension exact match
    AND (extension = $2 OR $2 IS NULL)
    -- File size range search
    AND (size >= $3 OR $3 IS NULL)
    AND (size <= $4 OR $4 IS NULL)
    -- File created_at range search
    AND (created_at >= $5 OR $5 IS NULL)
    AND (created_at <= $6 OR $6 IS NULL)
    -- File modified_At range search
    AND (modified_at >= $7 OR $7 IS NULL)
    AND (modified_at <= $8 OR $8 IS NULL)
    -- File accessed_at range search
    AND (accessed_at >= $9 OR $9 IS NULL)
    AND (accessed_at <= $10 OR $10 IS NULL)
    -- -- File content search
    AND (content ILIKE '%' || $11 || '%' OR $11 IS NULL)
OFFSET $12
-- If limit is not provided, return all results
LIMIT CASE WHEN $13 = 0 THEN NULL ELSE $12 END;