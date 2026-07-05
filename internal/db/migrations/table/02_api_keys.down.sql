-- Down Migration: Drop api_keys table and indexes

DROP INDEX CONCURRENTLY IF EXISTS idx_api_keys_revoked_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_api_keys_expires_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_api_keys_key_hash;
DROP INDEX CONCURRENTLY IF EXISTS idx_api_keys_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_api_keys_state;

DROP TABLE IF EXISTS api_keys;
