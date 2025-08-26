ALTER TABLE "mm_flow_step" DROP CONSTRAINT IF EXISTS "fk_mm_flow";
ALTER TABLE "mm_flow_step" DROP CONSTRAINT IF EXISTS "fk_mm_use_case_step";
ALTER TABLE "mm_flow_step" DROP CONSTRAINT IF EXISTS "fk_mm_use_case";

DROP TABLE IF EXISTS mm_flow_step;

