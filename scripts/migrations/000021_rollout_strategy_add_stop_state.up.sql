CREATE TYPE mm_rollout_state_new AS ENUM (
    'INIT',
    'WARMUP',
    'ESCAPED',
    'ADAPTIVE',
    'COMPLETED',
    'FORCED_STOP',
    'FORCED_ESCAPED',
    'FORCED_COMPLETED'
);

ALTER TABLE "mm_rollout_strategy" ALTER COLUMN rollout_state TYPE "mm_rollout_state_new" USING rollout_state::text::"mm_rollout_state_new";

DROP TYPE "mm_rollout_state";

ALTER TYPE "mm_rollout_state_new" RENAME TO "mm_rollout_state";