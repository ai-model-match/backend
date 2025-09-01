ALTER TABLE "mm_picker_request" DROP CONSTRAINT IF EXISTS "fk_mm_picker_request_use_case";
ALTER TABLE "mm_picker_request" DROP CONSTRAINT IF EXISTS "fk_mm_picker_request_use_case_step_id";
ALTER TABLE "mm_picker_request" DROP CONSTRAINT IF EXISTS "fk_mm_picker_request_flow_id";
ALTER TABLE "mm_picker_request" DROP CONSTRAINT IF EXISTS "fk_mm_picker_request_flow_step_id";

DROP TABLE "mm_picker_request";
