alter table "public"."usage_events" add column "lecture_id" uuid;

alter table "public"."usage_events" add constraint "usage_events_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES lectures(id) ON DELETE SET NULL not valid;

alter table "public"."usage_events" validate constraint "usage_events_lecture_id_fkey";


