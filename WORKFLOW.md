# Tech Stack

| Layer / Component                        | Technology & Notes                          |
| ---------------------------------------- | ------------------------------------------- |
| **Frontend**                             | Next.js (React)                             |
| • TipTap WYSIWYG editor                  |                                             |
| **API Gateway**                          | Golang (stlib)                              |
| • Route groups under `/v1` (see below)   |                                             |
| • JWT middleware validating Supabase JWT |                                             |
| **Auth**                                 | Supabase Auth (Google provider)             |
| **Object Storage**                       | Supabase Storage                            |
| **Relational & Vector**                  | Supabase Postgres (serverless)              |
| • pgvector (vector embeddings)           |                                             |
| **Message Queue**                        | Google Cloud Pub/Sub                        |
| **AI Microservices**                     | Python (FastAPI)                            |
| **PDF Parsing**                          | PyMuPDF                                     |
| **Embeddings**                           | OpenAI                                      |
| **LLM Inference**                        | OpenAI                                      |
| **Containerization**                     | Docker on Google Cloud Run                  |
| **CI/CD**                                | GitHub Actions                              |
| **Monitoring & Logging**                 | Supabase Logs / Sentry                      |
| **Cache (Later)**                        | Managed Redis (e.g. Upstash or Redis Cloud) |

# Repos

| Purpose                         | Type   | Deployment                    |
| ------------------------------- | ------ | ----------------------------- |
| Frontend                        | NextJS | Vercel Serverless             |
| Backend API Gateway             | Go     | Google Cloud Run (Serverless) |
| PDF processing and AI LLM calls | Python | Google Cloud Run (Serverless) |

# Pub/Sub Push-Based Workflow

We use Google Cloud Pub/Sub with push subscriptions to Python API endpoints:

- **Topics**: ingestion, embedding, explanation, summary
- **Subscriptions**: configured as push to `/{topic}` on your API server
- **Retry & Dead-Letter**: Each subscription has an exponential backoff policy (min:10s, max:10m). After exceeding max delivery attempts, failed messages are forwarded to a dead-letter topic, which pushes via HTTP POST to the `/dlq` endpoint on your API gateway. There, payloads are persisted in the database for logging and manual inspection.
- **Ack Deadline**: Configure each subscription's `ackDeadlineSeconds` to match your expected processing time (e.g., 60s), and use the client library's ack-deadline lease-extension API in long-running handlers to renew the deadline before it expires, preventing premature redelivery.

# Push Handlers (Cloud Run)

/ingest → Python ingestion endpoint (FastAPI)
/embedding → Python embedding endpoint (FastAPI)
/explanation → Python explanation endpoint (FastAPI)
/summary → Python summary endpoint (FastAPI)

Pub/Sub pushes directly to your Python services, which handle the entire async pipeline including status updates and publishes for downstream jobs.

# Key API Route Groups

```
/api/v1/courses
├── POST / → create course
├── GET /:courseId → fetch course
├── PATCH /:courseId → update course
└── DELETE /:courseId → delete course

/api/v1/lectures
├── POST / → create lecture
├── GET / → list lectures (query by course_id) (`?limit=&offset=`)
├── GET /:lectureId → fetch lecture
├── PATCH /:lectureId → update lecture metadata
└── DELETE /:lectureId → delete lecture

/api/v1/lectures/:lectureId
├── GET /summary → get lecture summary
├── GET /explanations → list lecture explanations (`?limit=&offset=`)
├── GET /notes → get lecture notes
├── POST /notes → create lecture note
└── PATCH /notes → update lecture note

/api/v1/users/me
├── GET / → fetch user profile
├── POST / → create or update profile
├── GET /courses → list user's courses
└── GET /recents → list recent lectures (`?limit=&offset=`)
```

# Authentication

1. **Sign-in**
   - Next.js → Supabase Auth (Google OAuth) → issues a JWT.
   - JWT stored in a secure, HTTP-only cookie.
2. **API Gateway**
   - Every request to `/api/v1/*` carries the Supabase JWT.
   - Go middleware verifies token, enforces row-level security on `user_id`.

# Data Flow

## 3.1. Client Upload → Go API

1. **Request**
   ```
   POST /api/v1/lectures
   Content-Type: multipart/form-data
   Body: { file: <PDF>, metadata… }
   ```
2. **Go API Handler**
   - Create lecture record with status `uploading`
   - Store PDF in Supabase Storage at `lectures/{lectureId}/original.pdf`
   - Persist the storage path to the database
   - Update lecture status to `pending_processing`
   - **Publish** a message to Pub/Sub topic `ingestion` with payload:
     ```json
     { "lecture_id": "<uuid>", "storage_path": "<path>" }
     ```
   - On error: roll back DB changes and/or emit cleanup job, return 500 to client

## 3.2. Ingestion Workflow (FastAPI)

**Trigger:** Pub/Sub pushes HTTP POST to `/ingest` with a JSON body like:

```json
{
  "message": {
    "messageId": "...",
    "publishTime": "...",
    "attributes": {
      /* custom attributes */
    },
    "data": "<Base64-encoded JSON payload>"
  },
  "subscription": "<subscriptionName>"
}
```

1. **Decode envelope with OIDC authentication**

   - Configure your push subscription with `pushConfig.oidcToken.serviceAccountEmail` set to your Cloud Run service account and `pushConfig.oidcToken.audience` set to your service URL.
   - Verify the OIDC JWT provided by Pub/Sub (issuer `https://pubsub.googleapis.com/`, audience matching your service URL) and validate expiry.
   - Base64-decode the `message.data` field and parse the resulting JSON into `{ lecture_id, storage_path }`, ensuring the payload does not exceed Pub/Sub's 10 MiB message size limit.

2. **Update lecture status**

   - Update lecture status to `parsing`

3. **Check for existing processing**: Before proceeding, check the lecture status in the database. If the status is already `embedding` (indicating ingestion is complete), skip all processing and return success immediately. If the status is `pending_processing` or `parsing`, proceed with processing.

4. **Resume from last successful point**: Query the database to find which slides have already been fully processed (have both text chunks and image records). Skip these slides and resume processing from the first unprocessed slide.

5. **Initialization**: The service initializes clients for Supabase Storage (S3) and Postgres, and optionally loads the Salesforce BLIP model for image captioning if its dependencies are installed.

6. **Download & Parse PDF**: It downloads the PDF from storage and opens it in memory using PyMuPDF. It then updates the lecture record in the database with the total number of slides.

7. **Process Each Slide**: The service iterates through each slide of the PDF.

   - **Text processing**:
     - **Check text processing**: Before processing each slide, check if it has already been fully processed (text chunks exist). If so, move on to the next step.
     - **Text Processing**: It extracts all raw text from the slide. This text is then broken down into smaller, overlapping chunks using the `tiktoken` library. Each chunk is saved to the database, and its ID is collected into a list for batch processing.
   - **Image processing**:
     - **Check image processing**: Before processing each slide, check if image records already exist. If so, move on to the next step.
     - **Image Processing**: It extracts all embedded images from the slide. For each image, it performs several steps:
     - **Analysis**: It runs Optical Character Recognition (OCR) using Tesseract and, if enabled, generates a descriptive caption (alt-text) using the BLIP model.
     - **Classification**: Based on keywords in the caption and the amount of text from OCR, it classifies the image as either "content" (e.g., diagrams, charts) or "decorative" (e.g., logos, backgrounds).
     - **Deduplication & Storage**: It computes a perceptual hash of each image to avoid storing duplicates. Decorative images are checked against a global table and stored in a shared `global/` folder if new. Content images are checked against a lecture-specific registry and stored in a folder for that lecture.
   - **Full Slide Rendering**: Finally, it renders a high-resolution image of the entire slide. This rendered image is also processed with OCR and BLIP, and the result is saved to storage. All image metadata (paths, hashes, OCR/alt-text) is stored in the database.
     - This image is then sent to Llama for captioning.

8. **Publish Embedding Job**: After processing all slides, the service publishes a single message to the `embedding` topic containing a list of all collected `chunk_id`s.

9. **Completion**: Once the embedding job is published, the service logs the completion of the ingestion task.

- On success:
  - Update lecture status to `embedding`
  - Return 200 (Pub/Sub acks)
- On failure:
  - Update lecture status to `failed` with error details
  - Return 500 (Pub/Sub retries then DLQ)

## 3.3. Embedding Push Handler (/embedding)

**Trigger:** Pub/Sub pushes HTTP POST to `/embedding`

1. **Decode envelope with OIDC authentication**

   - Verify the OIDC JWT provided by Pub/Sub (issuer `https://pubsub.googleapis.com/`, audience matching your service URL).
   - Base64-decode the `message.data` field and parse the JSON into `{ "chunk_ids": [...], lecture_id: <uuid> }`.

2. **Filter Processed Chunks**: For idempotency, the handler queries the database to find which of the received `chunk_id`s already have an embedding. It removes these from the list to avoid reprocessing. If no chunks remain, the handler returns success immediately.

3. **Fetch & Embed (Batch)**: It fetches the text for all remaining (unprocessed) chunks from the database. It then calls the OpenAI embedding API with the entire array of texts, receiving all vector embeddings in a single API call.

4. **Store Embeddings (Batch)**: The generated vectors are saved to the `embeddings` table in a single database transaction or bulk insert operation.

5. **Publish Explanation Jobs & Finalize**: After the batch embedding is complete, the service queries for all unique `slide_id`s associated with the lecture. For each slide, it publishes a separate message to the `explanation` topic. After enqueuing all explanation jobs, it updates the lecture's status to `explaining`.

6. **Completion**: The service logs the completion of the batch embedding task.

- On success:
  - Return 200 (Pub/Sub acks)
- On failure:
  - Update lecture status to `failed` with error details
  - Return 500 (Pub/Sub retries then DLQ)

## 3.4. Explanation Push Handler (/explanation)

**Trigger:** Pub/Sub pushes HTTP POST to `/explanation`

1. **Decode envelope with OIDC authentication**

   - Verify the OIDC JWT provided by Pub/Sub (issuer `https://pubsub.googleapis.com/`, audience matching your service URL).
   - Base64-decode the `message.data` field and parse the JSON into `{ slide_id, lecture_id, slide_number }`.

2. **Check for existing explanation**: Before processing, check if an explanation already exists for this specific slide in the database. If an explanation record already exists for this slide_id, skip all processing and return success immediately.

3. **Gather Context**:

   - **Recent History**: It fetches the short, one-liner summaries from the last 1-3 slides to understand the immediate context.
   - **Current Slide Data**: It retrieves the full text and any OCR/alt-text from images on the current slide.
   - **Related Concepts (RAG)**: It creates an embedding from the current slide's full text and uses it to perform a vector similarity search across the entire lecture. This retrieves the most relevant text chunks from other slides, providing broad, lecture-wide context.

4. **Prompt Assembly & LLM Call**: All the gathered information—recent history, current slide data, and related concepts—is assembled into a detailed prompt. It instructs the LLM to act as an AI professor and generate a clear, in-depth explanation. The LLM is asked to classify the slide's purpose (e.g., "cover", "header", "content") and return the output as a structured JSON object containing a `one_liner` summary and the full `content` in Markdown.

5. **Persist Explanation & Update Progress**: The generated explanation and one-liner are saved to the `explanations` table. The service then atomically increments the `processed_slides` counter for the lecture within a transaction to avoid lost updates.

6. **Publish Summary Job**: If all slides for the lecture have been explained (`processed_slides` equals `total_slides`), it updates the lecture's status to `summarising` and publishes a final message to the `summary` topic.

7. **Completion**: Once the explanation is complete, the service logs the completion of the explanation task.

- On success:
  - Return 200 (Pub/Sub acks)
- On failure:
  - Update lecture status to `failed` with error details
  - Return 500 (Pub/Sub retries then DLQ)

## 3.5. Summary Push Handler (/summary)

**Trigger:** Pub/Sub pushes HTTP POST to `/summary`

1. **Decode envelope with OIDC authentication**

   - Verify the OIDC JWT provided by Pub/Sub (issuer `https://pubsub.googleapis.com/`, audience matching your service URL).
   - Base64-decode the `message.data` field and parse the JSON into `{ lecture_id }`.

2. **Check for existing summary**: Before processing, check if a summary already exists for this specific lecture in the database. If a summary record already exists for this lecture_id, skip all processing and return success immediately.

3. **Gather All Explanations**: It retrieves all the detailed, slide-by-slide explanations from the database for the entire lecture.

4. **Build Prompt & Call LLM**: It combines all the explanations into a single, comprehensive prompt. It instructs the LLM to act as an AI professor and synthesize the information into a student-friendly "cheatsheet." The cheatsheet should start with a list of key takeaways and then provide a well-structured summary of the lecture's main topics.

5. **Persist Summary & Finalize Lecture**: The generated Markdown summary is saved to the `summaries` table. The service then updates the main lecture's status to `complete` and records a `completed_at` timestamp, marking the successful end of the entire processing pipeline.

6. **Metrics & Errors**: log token usage, cost, and fallback on over-length.

7. **Completion**: Once the summary is complete, the service logs the completion of the summary task.

- On success:
  - Return 200 (Pub/Sub acks)
- On failure:
  - Update lecture status to `failed` with error details
  - Return 500 (Pub/Sub retries then DLQ)
