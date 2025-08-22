-- Create index for fuzzy search
CREATE INDEX idx_mm_use_case_search ON mm_use_case
USING GIN 
(
	to_tsvector('simple', 
    regexp_replace(mm_use_case.code, '[^\w]+', ' ', 'g') || ' ' || 
    regexp_replace(mm_use_case.title, '[^\w]+', ' ', 'g') || ' ' || 
    regexp_replace(mm_use_case.description, '[^\w]+', ' ', 'g') || ' ' || 
    mm_use_case.code || ' ' || 
    mm_use_case.title || ' ' || 
    mm_use_case.description
  )
);