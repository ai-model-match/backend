CREATE TYPE "mm_rollout_state" AS ENUM (
  'INIT',
  'REVIEW',
  'WARMUP',
  'ESCAPED',
  'MONITOR',
  'ADAPTIVE',
  'COMPLETED',
  'FORCED_ESCAPED',
  'FORCED_COMPLETED'
);

CREATE TABLE "mm_rollout_strategy" (
    "id" VARCHAR(36) PRIMARY KEY,
    "use_case_id" VARCHAR(36) NOT NULL,
    "rollout_state" mm_rollout_state NOT NULL,
    "configuration" JSON NOT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL
);

ALTER TABLE "mm_rollout_strategy"
    ADD CONSTRAINT "fk_mm_rollout_strategy_use_case"
    FOREIGN KEY ("use_case_id") REFERENCES mm_use_case(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;