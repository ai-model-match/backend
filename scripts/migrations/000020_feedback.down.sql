ALTER TABLE "mm_feedback" DROP CONSTRAINT IF EXISTS "fk_mm_feedback_use_case";
ALTER TABLE "mm_feedback" DROP CONSTRAINT IF EXISTS "fk_mm_feedback_flow";

DROP TABLE IF EXISTS "mm_feedback";