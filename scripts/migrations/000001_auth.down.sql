ALTER TABLE "mm_auth_session" DROP CONSTRAINT "idx_mm_auth_session_refresh_token";

DROP TABLE IF EXISTS "mm_auth_session";

DROP EXTENSION IF EXISTS pg_trgm;