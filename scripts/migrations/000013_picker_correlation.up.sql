CREATE TABLE "mm_picker_correlation" (
    "id" VARCHAR(36) PRIMARY KEY,
    "use_case_id" VARCHAR(36) NOT NULL,
    "flow_id" VARCHAR(36) NOT NULL,
    "fallback" boolean NOT NULL,
    "created_at" TIMESTAMP NOT NULL
);

ALTER TABLE "mm_picker_correlation"
    ADD CONSTRAINT "fk_mm_picker_correlation_use_case"
    FOREIGN KEY ("use_case_id") REFERENCES mm_use_case(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE "mm_picker_correlation"
    ADD CONSTRAINT "fk_mm_picker_correlation_flow"
    FOREIGN KEY ("flow_id") REFERENCES mm_flow(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;


ALTER TABLE "mm_picker_correlation" ADD CONSTRAINT "idx_mm_picker_correlation_use_case" UNIQUE ("use_case_id");

CREATE INDEX "idx_mm_picker_correlation_use_case_id_created_at" ON "mm_picker_correlation" ("use_case_id", "created_at");
