CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE "mm_auth_session" (
    "id" varchar(36) PRIMARY KEY NOT NULL,
    "username" varchar(255) NOT NULL,
    "created_at" timestamp NOT NULL,
    "expires_at" timestamp NOT NULL,
    "refresh_token" text NOT NULL
);

ALTER TABLE "mm_auth_session" ADD CONSTRAINT "idx_mm_auth_session_refresh_token" UNIQUE ("refresh_token");