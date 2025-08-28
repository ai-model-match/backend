CREATE TABLE "mm_flow_statistics" (
    "id" VARCHAR(36) PRIMARY KEY,
    "flow_id" VARCHAR(36) NOT NULL,
    "use_case_id" VARCHAR(36) NOT NULL,
    "tot_req" BIGINT NOT NULL,
    "tot_sess_req" BIGINT NOT NULL,
    "initial_pct" DOUBLE PRECISION NOT NULL,
    "current_pct" DOUBLE PRECISION NOT NULL,
    "tot_feedback" BIGINT NOT NULL,
    "avg_feedback_score" DOUBLE PRECISION NOT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL
);

ALTER TABLE "mm_flow_statistics"
    ADD CONSTRAINT "fk_mm_flow_statistics_flow"
    FOREIGN KEY ("flow_id") REFERENCES mm_flow(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE "mm_flow_statistics"
    ADD CONSTRAINT "fk_mm_flow_statistics_use_case"
    FOREIGN KEY ("use_case_id") REFERENCES mm_use_case(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;