-- Seed data for Supabase schema

-- 1. Insert a user into auth.users (replace UUIDs as needed)
INSERT INTO auth.users (id, email, encrypted_password)
VALUES
('00000000-0000-0000-0000-000000000001', 'user1@example.com', 'password1');

-- Log users table before inserting into user_profiles
SELECT * FROM auth.users;

-- 2. User Profile
INSERT INTO user_profiles (user_id, name, email, avatar_url)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'Alice Example', 'user1@example.com', 'https://example.com/avatar1.png');

-- 3. Courses
INSERT INTO courses (id, user_id, title, description, is_default)
VALUES
  ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Drafts', 'A beginner course on AI', TRUE),
  ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Intro to AI', 'A beginner course on AI', FALSE);

-- 4. Lectures
INSERT INTO lectures (id, user_id, course_id, title, pdf_url, status)
VALUES
  ('20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'Lecture 1: What is AI?', 'https://example.com/pdf/lecture1.pdf', 'uploaded');

-- 5. Slides
INSERT INTO slides (id, lecture_id, slide_number)
VALUES
  ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 1);

-- 6. Summaries
INSERT INTO summaries (lecture_id, content)
VALUES
  ('20000000-0000-0000-0000-000000000001', 'This lecture introduces the basics of AI.');

-- 7. Explanations
INSERT INTO explanations (id, lecture_id, slide_number, content)
VALUES
  ('50000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 1, 'AI is the simulation of human intelligence by machines.');

-- 8. Slide Images
INSERT INTO slide_images (id, lecture_id, slide_number, image_index, storage_path, caption, width, height)
VALUES
  ('60000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 1, 1, '/slides/ai_intro/slide1.png', 'AI Concept', 800, 600);

-- 9. Notes
INSERT INTO notes (id, user_id, lecture_id, content)
VALUES
  ('70000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Remember to review the definition of AI.');

-- 10. Create Queues
SELECT pgmq.create('ingestion_queue');
SELECT pgmq.create('embedding_queue');
SELECT pgmq.create('explanation_queue');
SELECT pgmq.create('summary_queue');
