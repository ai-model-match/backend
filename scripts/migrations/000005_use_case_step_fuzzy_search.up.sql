-- Create index for fuzzy search
CREATE INDEX idx_mm_use_case_step_search ON mm_use_case_step
USING GIN 
(
	to_tsvector('simple', 
    regexp_replace(mm_use_case_step.code, '[^\w]+', ' ', 'g') || ' ' || 
    regexp_replace(mm_use_case_step.title, '[^\w]+', ' ', 'g') || ' ' || 
    regexp_replace(mm_use_case_step.description, '[^\w]+', ' ', 'g') || ' ' || 
    mm_use_case_step.code || ' ' || 
    mm_use_case_step.title || ' ' || 
    mm_use_case_step.description
  )
);