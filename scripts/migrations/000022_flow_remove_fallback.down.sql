ALTER TABLE "mm_flow" ADD COLUMN "fallback" BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE "mm_picker_correlation" ADD COLUMN "fallback" BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE "mm_picker_request" ADD COLUMN "is_fallback" BOOLEAN NOT NULL DEFAULT false;
