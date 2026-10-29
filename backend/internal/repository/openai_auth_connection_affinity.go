package repository

import "context"

func (r *openAIOAuthReauthRepository) ResolveAuthConnectionAffinity(ctx context.Context, accountID, apiKeyID int64, sessionHash, preferred string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `WITH expired AS (
      DELETE FROM openai_auth_connection_affinity WHERE ctid IN (SELECT ctid FROM openai_auth_connection_affinity WHERE expires_at<NOW() LIMIT 1000)
    ) INSERT INTO openai_auth_connection_affinity(account_id,api_key_id,session_hash,connection_id,expires_at)
    VALUES($1,$2,$3,$4,NOW()+INTERVAL '1 hour') ON CONFLICT(account_id,api_key_id,session_hash) DO UPDATE
    SET connection_id=CASE WHEN openai_auth_connection_affinity.expires_at>NOW() THEN openai_auth_connection_affinity.connection_id ELSE EXCLUDED.connection_id END,
        expires_at=EXCLUDED.expires_at RETURNING connection_id::text`, accountID, apiKeyID, sessionHash, preferred).Scan(&id)
	return id, err
}
