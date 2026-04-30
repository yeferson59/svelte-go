DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token;

DROP INDEX IF EXISTS idx_accounts_provider_id;
DROP INDEX IF EXISTS idx_accounts_account_id;
DROP INDEX IF EXISTS idx_accounts_user_id;

DROP INDEX IF EXISTS idx_verifications_identifier;
DROP INDEX IF EXISTS idx_verifications_value;


DROP TABLE IF EXISTS accounts CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS verifications CASCADE;
