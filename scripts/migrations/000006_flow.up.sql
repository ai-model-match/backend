CREATE TABLE "mm_flow" (
    "id" VARCHAR(36) PRIMARY KEY,
    "use_case_id" VARCHAR(36) NOT NULL,
    "title" VARCHAR(255) NOT NULL,
    "description" TEXT,
    "active" BOOLEAN NOT NULL DEFAULT false,
    "fallback" BOOLEAN NOT NULL DEFAULT false,
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL
);

ALTER TABLE "mm_flow" 
ADD CONSTRAINT "fk_mm_flow_use_case" 
FOREIGN KEY ("use_case_id") REFERENCES mm_use_case(id)
ON UPDATE CASCADE
ON DELETE CASCADE;