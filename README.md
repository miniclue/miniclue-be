**Base & Versioning**

* **Base path:** `/api/v1`
* Bump to `/api/v2` on breaking changes

---

**Courses**

* `POST   /api/v1/courses`
  * Create a new course (`{ title, description?, isDefault?: boolean }`)
* `GET    /api/v1/courses/{courseId}`
  * Fetch metadata & status for one course
* `PUT    /api/v1/courses/{courseId}`
  * Update title/description/etc.
* `DELETE /api/v1/courses/{courseId}`
  * Delete a course and all its lectures

---

**Lectures**

* `GET    /api/v1/courses/{courseId}/lectures`
  * List lectures in that course (`?status=&limit=&offset=`)
* `POST   /api/v1/courses/{courseId}/lectures/upload`
  * Upload PDF + metadata → creates a new lecture under that course (or default drafts course if using `/courses/default/lectures/upload`) and returns `{ lectureId }`
* `GET    /api/v1/courses/{courseId}/lectures/{lectureId}`
  * Fetch lecture metadata, status & PDF
* `PUT    /api/v1/courses/{courseId}/lectures/{lectureId}`
  * Update lecture metadata (title, tags, etc.)
* `DELETE /api/v1/courses/{courseId}/lectures/{lectureId}`
  * Delete a lecture and all its derived data

---

**Summary**

* `GET /api/v1/courses/{courseId}/lectures/{lectureId}/summary`
  * Retrieve (or trigger+retrieve) TLDR summary; responds with `Cache-Control: max-age=<TTL>`

---

**Slides & Explanations**

* `GET /api/v1/courses/{courseId}/lectures/{lectureId}/explanations`
  * Use limit and offset params
  * Cached in Redis

---

**Notes**

* `GET    /api/v1/courses/{courseId}/lectures/{lectureId}/notes`
  * List user-saved notes (`?limit=&offset=`)
* `POST   /api/v1/courses/{courseId}/lectures/{lectureId}/notes`
  * Create new note (`{ content: string }`)
* `PUT    /api/v1/courses/{courseId}/lectures/{lectureId}/notes/{noteId}`
  * Update existing note
* `DELETE /api/v1/courses/{courseId}/lectures/{lectureId}/notes/{noteId}`
  * Delete a note

---

**Users & Dashboard**

* `GET  /api/v1/users/me`
  * Fetch current user’s profile for dashboard
* `POST /api/v1/users/me`
  * Create a new user profile
* `PUT  /api/v1/users/me` (to be implemented)
  * Update own profile (name, avatar, prefs)
* `GET  /api/v1/users/me/courses`
  * List courses the user owns or is enrolled in
* `GET  /api/v1/users/me/recents`
  * List the user’s most recently created or accessed lectures (`?limit=&offset=`)


**Filtering, Pagination & Caching**

* **Filtering:** via query params (e.g. `?status=parsed`)
* **Pagination:** `?limit=&offset=` (or `?page=&pageSize=`)
* **Cache-Control:** summary & explanation endpoints include `Cache-Control: max-age=<TTL>`; clients revalidate after TTL
