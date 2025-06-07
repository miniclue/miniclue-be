-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS vector;    -- pgvector
CREATE EXTENSION IF NOT EXISTS pgmq;

SET search_path TO public;

-------------------------------------------------------------------------------
-- 1. Course Table
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS courses (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  title       TEXT        NOT NULL,
  description TEXT        DEFAULT '',
  is_default  BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_courses_user_id ON courses(user_id);


-------------------------------------------------------------------------------
-- 2. User Profile Table
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_profiles (
  user_id    UUID        PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  full_name  TEXT        DEFAULT '',
  avatar_url TEXT        DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-------------------------------------------------------------------------------
-- 3. Lecture Table
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS lectures (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID        NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  course_id  UUID        NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  title      TEXT        NOT NULL,
  pdf_url    TEXT        NOT NULL,
  status     VARCHAR(32) NOT NULL DEFAULT 'uploaded',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_lectures_user_id   ON lectures(user_id);
CREATE INDEX IF NOT EXISTS idx_lectures_course_id ON lectures(course_id);


-------------------------------------------------------------------------------
-- 4. Slide Table
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS slides (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  lecture_id   UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT         NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(lecture_id, slide_number)
);
CREATE INDEX IF NOT EXISTS idx_slides_lecture_id        ON slides(lecture_id);
CREATE INDEX IF NOT EXISTS idx_slides_lecture_slide     ON slides(lecture_id, slide_number);


-------------------------------------------------------------------------------
-- 5. Chunk Table
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS chunks (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  lecture_id   UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT         NOT NULL,
  chunk_index  INT         NOT NULL,
  text         TEXT        NOT NULL,
  is_image     BOOLEAN     NOT NULL DEFAULT FALSE,
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
CREATE INDEX IF NOT EXISTS idx_chunks_lecture       ON chunks(lecture_id);


-------------------------------------------------------------------------------
-- 6. Embedding Table (pgvector)
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS embeddings (
  chunk_id     UUID         PRIMARY KEY,
  lecture_id   UUID         NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT          NOT NULL,
  vector       VECTOR(1536) NOT NULL,
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
ALTER TABLE embeddings
  ADD CONSTRAINT fk_embeddings_chunk
    FOREIGN KEY (chunk_id)
      REFERENCES chunks (id)
      ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_embeddings_vector       ON embeddings USING ivfflat (vector) WITH (lists = 100);
CREATE INDEX IF NOT EXISTS idx_embeddings_lecture_snap ON embeddings(lecture_id, slide_number);


-------------------------------------------------------------------------------
-- 7. Summary Table
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS summaries (
  lecture_id  UUID        PRIMARY KEY REFERENCES lectures(id) ON DELETE CASCADE,
  content     TEXT        NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_summaries_by_lecture ON summaries (lecture_id);


-------------------------------------------------------------------------------
-- 8. Explanation Table
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS explanations (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  lecture_id   UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT         NOT NULL,
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


-------------------------------------------------------------------------------
-- 9. Slide Images Table
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS slide_images (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  lecture_id   UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  slide_number INT         NOT NULL,
  image_index  INT         NOT NULL,
  storage_path TEXT        NOT NULL,
  caption      TEXT        NOT NULL DEFAULT '',
  width        INT         NOT NULL,
  height       INT         NOT NULL,
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


-------------------------------------------------------------------------------
-- 10. Note Table
-------------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notes (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID        NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  lecture_id UUID        NOT NULL REFERENCES lectures(id) ON DELETE CASCADE,
  content    TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_notes_user_id    ON notes(user_id);
CREATE INDEX IF NOT EXISTS idx_notes_lecture_id ON notes(lecture_id);
