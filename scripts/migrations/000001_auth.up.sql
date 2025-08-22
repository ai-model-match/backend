CREATE TABLE "mm_auth" (
    "id" varchar(36) PRIMARY KEY NOT NULL,
    "username" varchar(255) NOT NULL,
    "created_at" timestamp NOT NULL,
    "expires_at" timestamp NOT NULL,
    "refresh_token" text NOT NULL
);

ALTER TABLE "mm_auth" ADD CONSTRAINT "idx_mm_auth_refresh_token" UNIQUE ("refresh_token");