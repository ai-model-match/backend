ALTER TABLE "mm_flow_statistics" DROP CONSTRAINT IF EXISTS "fk_mm_flow_statistics_flow";
ALTER TABLE "mm_flow_statistics" DROP CONSTRAINT IF EXISTS "fk_mm_flow_statistics_use_case";

DROP TABLE IF EXISTS "mm_flow_statistics";