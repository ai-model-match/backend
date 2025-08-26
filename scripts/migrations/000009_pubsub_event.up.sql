CREATE TABLE "mm_event" (
    "id" VARCHAR(36) PRIMARY KEY,
    "topic" VARCHAR(255) NOT NULL,
    "event_type" VARCHAR(255) NOT NULL,
    "event_date" TIMESTAMP NOT NULL,
    "event_body" JSON NOT NULL
);

-- Composite index for fast lookups by topic + date range
CREATE INDEX idx_mm_event_topic_event_date ON "mm_event" ("topic", "event_date");