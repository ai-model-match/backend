ALTER TABLE "mm_picker_correlation" DROP CONSTRAINT IF EXISTS "fk_mm_picker_correlation_use_case";
ALTER TABLE "mm_picker_correlation" DROP CONSTRAINT IF EXISTS "fk_mm_picker_correlation_flow";
ALTER TABLE "mm_picker_correlation" DROP CONSTRAINT IF EXISTS "idx_mm_picker_correlation_use_case";
DROP INDEX "idx_mm_picker_correlation_use_case_id_created_at";

DROP TABLE IF EXISTS "mm_picker_correlation";
