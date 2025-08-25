-- Create index for fuzzy search
CREATE INDEX idx_mm_flow_search ON mm_flow
USING GIN 
(
	to_tsvector('simple', 
    regexp_replace(mm_flow.title, '[^\w]+', ' ', 'g') || ' ' || 
    regexp_replace(mm_flow.description, '[^\w]+', ' ', 'g') || ' ' || 
    mm_flow.title || ' ' || 
    mm_flow.description
  )
);