CREATE TYPE mm_rollout_state_new AS ENUM (
    'INIT',
    'WARMUP',
    'ESCAPED',
    'ADAPTIVE',
    'COMPLETED',
    'FORCED_ESCAPED',
    'FORCED_COMPLETED'
);

UPDATE "mm_rollout_strategy" SET rollout_state = 'INIT' WHERE rollout_state = 'MONITOR';

ALTER TABLE "mm_rollout_strategy" ALTER COLUMN rollout_state TYPE "mm_rollout_state_new" USING rollout_state::text::"mm_rollout_state_new";

DROP TYPE "mm_rollout_state";

ALTER TYPE "mm_rollout_state_new" RENAME TO "mm_rollout_state";