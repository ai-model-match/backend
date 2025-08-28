CREATE TABLE "mm_flow_step" (
    "id" VARCHAR(36) PRIMARY KEY,
    "flow_id" VARCHAR(36) NOT NULL,
    "use_case_id" VARCHAR(36) NOT NULL,
    "use_case_step_id" VARCHAR(36) NOT NULL,
    "configuration" JSON NOT NULL,
    "placeholders" JSON NOT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL
);

ALTER TABLE "mm_flow_step"
    ADD CONSTRAINT "fk_mm_flow"
    FOREIGN KEY ("flow_id") REFERENCES mm_flow(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE "mm_flow_step"
    ADD CONSTRAINT "fk_mm_use_case"
    FOREIGN KEY ("use_case_id") REFERENCES mm_use_case(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE "mm_flow_step"
    ADD CONSTRAINT "fk_mm_use_case_step"
    FOREIGN KEY ("use_case_step_id") REFERENCES mm_use_case_step(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;