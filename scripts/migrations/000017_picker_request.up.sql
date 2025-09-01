CREATE TABLE "mm_picker_request" (
    "id" VARCHAR(36) PRIMARY KEY,
    "use_case_id" VARCHAR(36) NOT NULL,
    "use_case_step_id" VARCHAR(36) NOT NULL,
    "flow_id" VARCHAR(36) NOT NULL,
    "flow_step_id" VARCHAR(36) NOT NULL,
    "correlation_id" VARCHAR(36) NOT NULL,
    "is_first_correlation" BOOLEAN NOT NULL,
    "is_fallback" BOOLEAN NOT NULL,
    "input_message" JSON NOT NULL,
    "output_message" JSON NOT NULL,
    "placeholders" JSON NOT NULL,
    "created_at" TIMESTAMP NOT NULL
);

ALTER TABLE "mm_picker_request"
    ADD CONSTRAINT "fk_mm_picker_request_use_case"
    FOREIGN KEY ("use_case_id") REFERENCES mm_use_case(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE "mm_picker_request"
    ADD CONSTRAINT "fk_mm_picker_request_use_case_step_id"
    FOREIGN KEY ("use_case_step_id") REFERENCES mm_use_case_step(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE "mm_picker_request"
    ADD CONSTRAINT "fk_mm_picker_request_flow_id"
    FOREIGN KEY ("flow_id") REFERENCES mm_flow(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE "mm_picker_request"
    ADD CONSTRAINT "fk_mm_picker_request_flow_step_id"
    FOREIGN KEY ("flow_step_id") REFERENCES mm_flow_step(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;