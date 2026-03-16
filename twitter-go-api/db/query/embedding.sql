-- name: ListRelatedTweetsByEmbedding :many
SELECT tweet_id, content
FROM tweet_embeddings
WHERE embedding <-> $1::vector < 1.0
ORDER BY embedding <-> $1::vector ASC
LIMIT $2;
