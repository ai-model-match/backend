ALTER TABLE "mm_rollout_strategy" DROP CONSTRAINT IF EXISTS "fk_mm_rollout_strategy_use_case";

DROP TABLE IF EXISTS "mm_rollout_strategy";

DROP TYPE IF EXISTS "mm_rollout_state";
