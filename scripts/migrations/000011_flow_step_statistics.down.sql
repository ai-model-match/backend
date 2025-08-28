ALTER TABLE "mm_flow_step_statistics" DROP CONSTRAINT IF EXISTS "fk_mm_flow_step_statistics_flow_step";
ALTER TABLE "mm_flow_step_statistics" DROP CONSTRAINT IF EXISTS "fk_mm_flow_statistics_flow";


DROP TABLE IF EXISTS "mm_flow_step_statistics";