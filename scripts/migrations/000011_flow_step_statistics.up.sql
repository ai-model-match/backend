CREATE TABLE "mm_flow_step_statistics" (
    "id" VARCHAR(36) PRIMARY KEY,
    "flow_step_id" VARCHAR(36) NOT NULL,
    "flow_id" VARCHAR(36) NOT NULL,
    "tot_req" BIGINT NOT NULL,
    "tot_sess_req" BIGINT NOT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL
);

ALTER TABLE "mm_flow_step_statistics"
    ADD CONSTRAINT "fk_mm_flow_step_statistics_flow_step"
    FOREIGN KEY ("flow_step_id") REFERENCES mm_flow_step(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE "mm_flow_step_statistics"
    ADD CONSTRAINT "fk_mm_flow_statistics_flow"
    FOREIGN KEY ("flow_id") REFERENCES mm_flow(id)
    ON UPDATE CASCADE
    ON DELETE CASCADE;