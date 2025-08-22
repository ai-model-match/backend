CREATE TABLE "mm_use_case" (
    "id" varchar(36) PRIMARY KEY NOT NULL,
    "code" varchar(255) NOT NULL,
    "title" varchar(255) NOT NULL,
    "description" text,
    "active" boolean NOT NULL DEFAULT false,
    "created_at" timestamp NOT NULL,
    "updated_at" timestamp NOT NULL
);

ALTER TABLE "mm_use_case" ADD CONSTRAINT "idx_mm_use_case_code" UNIQUE ("code");
