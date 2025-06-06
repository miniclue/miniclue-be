-- Insert sample data for testing

-- Insert sample lectures
INSERT INTO lectures (id, user_id, title, status)
VALUES 
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'Sample Lecture 1', 'explained'),
  ('33333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 'Sample Lecture 2', 'parsed'),
  ('44444444-4444-4444-4444-444444444444', '22222222-2222-2222-2222-222222222222', 'Advanced Mathematics', 'explained'),
  ('55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 'Physics Fundamentals', 'parsed'),
  ('66666666-6666-6666-6666-666666666666', '22222222-2222-2222-2222-222222222222', 'Chemistry Basics', 'uploaded'),
  ('77777777-7777-7777-7777-777777777777', '22222222-2222-2222-2222-222222222222', 'Biology Overview', 'explained'),
  ('88888888-8888-8888-8888-888888888888', '22222222-2222-2222-2222-222222222222', 'Computer Science 101', 'parsed');


-- Insert sample slides
INSERT INTO slides (lecture_id, slide_number, image_keys)
VALUES 
  ('11111111-1111-1111-1111-111111111111', 1, ARRAY['slides/lecture1/slide1.jpg']),
  ('11111111-1111-1111-1111-111111111111', 2, ARRAY['slides/lecture1/slide2.jpg']),
  ('33333333-3333-3333-3333-333333333333', 1, ARRAY['slides/lecture2/slide1.jpg']);

-- Insert sample explanations
INSERT INTO explanations (lecture_id, slide_number, content)
VALUES 
  ('11111111-1111-1111-1111-111111111111', 1, 'This slide introduces the main topic and provides an overview of key concepts that will be covered in detail later.'),
  ('11111111-1111-1111-1111-111111111111', 2, 'Here we dive deeper into concept A, explaining its importance and applications.');

-- Insert sample summaries
INSERT INTO summaries (lecture_id, content)
VALUES 
  ('11111111-1111-1111-1111-111111111111', 'This lecture covers the fundamental concepts of the topic, with a focus on concept A and its practical applications.');

-- Insert sample notes
INSERT INTO notes (user_id, lecture_id, content)
VALUES 
  ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Important points to remember about concept A');

-- Insert sample slide images
INSERT INTO slide_images (lecture_id, slide_number, image_index, storage_path, caption, width, height)
VALUES 
  ('11111111-1111-1111-1111-111111111111', 1, 0, 'slides/lecture1/slide1.jpg', 'Introduction slide with key concepts', 1920, 1080),
  ('11111111-1111-1111-1111-111111111111', 2, 0, 'slides/lecture1/slide2.jpg', 'Detailed explanation of concept A', 1920, 1080);
