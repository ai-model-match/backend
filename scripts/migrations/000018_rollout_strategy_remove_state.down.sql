CREATE TYPE mm_rollout_state_old AS ENUM (
    'INIT',
    'MONITOR',
    'WARMUP',
    'ESCAPED',
    'ADAPTIVE',
    'COMPLETED',
    'FORCED_ESCAPED',
    'FORCED_COMPLETED'
);

ALTER TABLE "mm_rollout_strategy" ALTER COLUMN rollout_state TYPE "mm_rollout_state_old" USING rollout_state::text::"mm_rollout_state_old";

DROP TYPE "mm_rollout_state";

ALTER TYPE "mm_rollout_state_old" RENAME TO "mm_rollout_state";