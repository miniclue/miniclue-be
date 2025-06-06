-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pgmq;

SET search_path TO public;

-- *********************
-- 1. Lecture Table
-- *********************
CREATE TABLE IF NOT EXISTS lectures (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID            NOT NULL,           -- Supabase Auth user UUID
  title      TEXT            NOT NULL,
  status     VARCHAR(32)     NOT NULL DEFAULT 'uploaded',  -- e.g. 'uploaded', 'parsed', 'explained'
  created_at TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Optionally add index on user_id for querying a user’s lectures
CREATE INDEX IF NOT EXISTS idx_lectures_user_id ON lectures(user_id);


-- *********************
-- 2. Slide Table
-- *********************
CREATE TABLE IF NOT EXISTS slides (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  lecture_id   UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT         NOT NULL,             -- 1-based index
  image_keys   TEXT[]      NOT NULL DEFAULT '{}', -- array of Supabase Storage paths for images on this slide
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(lecture_id, slide_number)
);

CREATE INDEX IF NOT EXISTS idx_slides_lecture_id ON slides(lecture_id);
CREATE INDEX IF NOT EXISTS idx_slides_lecture_slide ON slides(lecture_id, slide_number);


-- *********************
-- 3. Chunk Table
-- *********************
CREATE TABLE IF NOT EXISTS chunks (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  lecture_id   UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT         NOT NULL,             -- which slide this chunk came from
  chunk_index  INT         NOT NULL,             -- 0-based index within that slide
  text         TEXT        NOT NULL,             -- raw text or image caption/OCR result
  is_image     BOOLEAN     NOT NULL DEFAULT FALSE, -- true if this chunk is from an image caption
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(lecture_id, slide_number, chunk_index)
);

ALTER TABLE chunks
  ADD CONSTRAINT fk_chunks_slide
    FOREIGN KEY (lecture_id, slide_number)
      REFERENCES slides (lecture_id, slide_number)
      ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_chunks_lecture_slide ON chunks(lecture_id, slide_number);
CREATE INDEX IF NOT EXISTS idx_chunks_lecture ON chunks(lecture_id);


-- *********************
-- 4. Embedding Table (pgvector)
-- *********************
CREATE TABLE IF NOT EXISTS embeddings (
  chunk_id     UUID               PRIMARY KEY,         -- matches chunks.id
  lecture_id   UUID               NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT                NOT NULL,            -- same as chunks.slide_number
  vector       VECTOR(1536)       NOT NULL,            -- size matches the embedding dimension (adjust as needed)
  created_at   TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

ALTER TABLE embeddings
  ADD CONSTRAINT fk_embeddings_chunk
    FOREIGN KEY (chunk_id)
      REFERENCES chunks (id)
      ON DELETE CASCADE;

-- Index for fast kNN searches on 'vector'
CREATE INDEX IF NOT EXISTS idx_embeddings_vector ON embeddings USING ivfflat (vector) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_embeddings_lecture_snap ON embeddings(lecture_id, slide_number);



-- *********************
-- 5. Summary Table
-- *********************
-- 5. Summary Table (versioned)
CREATE TABLE IF NOT EXISTS summaries (
  lecture_id  UUID        PRIMARY KEY REFERENCES lectures(id) ON DELETE CASCADE,
  content     TEXT        NOT NULL,                 -- cheat-sheet summary text
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index for quickly fetching all versions for a lecture
CREATE INDEX IF NOT EXISTS idx_summaries_by_lecture ON summaries (lecture_id);


-- *********************
-- 6. Explanation Table
-- *********************
CREATE TABLE IF NOT EXISTS explanations (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  lecture_id   UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT         NOT NULL,            -- which slide this explanation is for
  content      TEXT        NOT NULL,           
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(lecture_id, slide_number)
);

ALTER TABLE explanations 
  ADD CONSTRAINT fk_explanations_slide
    FOREIGN KEY (lecture_id, slide_number)
      REFERENCES slides (lecture_id, slide_number)
      ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_explanations_lecture_slide ON explanations(lecture_id, slide_number);


-- *********************
-- 7. SlideImage Table
-- *********************
CREATE TABLE IF NOT EXISTS slide_images (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  lecture_id   UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT         NOT NULL,            -- slide index
  image_index  INT         NOT NULL,            -- 0-based index for multiple images on a slide
  storage_path TEXT        NOT NULL,            -- Supabase Storage path to the extracted image
  caption      TEXT        NOT NULL DEFAULT '', -- generated caption or OCR text
  width        INT         NOT NULL,            -- image width in pixels
  height       INT         NOT NULL,            -- image height in pixels
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(lecture_id, slide_number, image_index)
);

ALTER TABLE slide_images
  ADD CONSTRAINT fk_slide_images_slide
    FOREIGN KEY (lecture_id, slide_number)
      REFERENCES slides (lecture_id, slide_number)
      ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_slide_images_lecture_slide ON slide_images(lecture_id, slide_number);


-- *********************
-- 8. Note Table
-- *********************
CREATE TABLE IF NOT EXISTS notes (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID        NOT NULL,               -- Supabase Auth user UUID
  lecture_id UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  content    TEXT        NOT NULL,               -- rich text (Markdown/HTML)
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id);
CREATE INDEX IF NOT EXISTS idx_notes_lecture_id ON notes(lecture_id);


