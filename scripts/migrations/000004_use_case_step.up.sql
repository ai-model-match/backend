CREATE TABLE "mm_use_case_step" (
    "id" VARCHAR(36) PRIMARY KEY,
    "use_case_id" VARCHAR(36) NOT NULL,
    "title" VARCHAR(255) NOT NULL,
    "code" VARCHAR(255) NOT NULL,
    "description" TEXT,
    "position" BIGINT NOT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL
);

ALTER TABLE "mm_use_case_step"
ADD CONSTRAINT "fk_mm_use_case"
FOREIGN KEY ("use_case_id") REFERENCES mm_use_case(id)
ON UPDATE CASCADE
ON DELETE CASCADE;

ALTER TABLE "mm_use_case_step" ADD CONSTRAINT "idx_mm_use_case_step_code" UNIQUE ("use_case_id", "code");