CREATE TABLE "mm_feedback" (
    "id" VARCHAR(36) PRIMARY KEY,
    "use_case_id" VARCHAR(36),
    "flow_id" VARCHAR(36),
    "correlation_id" VARCHAR(36),
    "score" DOUBLE PRECISION,
    "comment" TEXT,
    "created_at" TIMESTAMP NOT NULL
);

ALTER TABLE "mm_feedback"
    ADD CONSTRAINT "fk_mm_feedback_use_case"
    FOREIGN KEY ("use_case_id") REFERENCES mm_use_case(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE "mm_feedback"
    ADD CONSTRAINT "fk_mm_feedback_flow"
    FOREIGN KEY ("flow_id") REFERENCES mm_flow(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;
