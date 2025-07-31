create type "public"."usage_event_type" as enum ('lecture_upload');

create table "public"."usage_events" (
    "id" uuid not null default gen_random_uuid(),
    "user_id" uuid not null,
    "event_type" usage_event_type not null,
    "created_at" timestamp with time zone not null default now()
);


alter table "public"."usage_events" enable row level security;

CREATE INDEX idx_usage_events_user_event_time ON public.usage_events USING btree (user_id, event_type, created_at);

CREATE UNIQUE INDEX usage_events_pkey ON public.usage_events USING btree (id);

alter table "public"."usage_events" add constraint "usage_events_pkey" PRIMARY KEY using index "usage_events_pkey";

alter table "public"."usage_events" add constraint "usage_events_user_id_fkey" FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE not valid;

alter table "public"."usage_events" validate constraint "usage_events_user_id_fkey";

grant delete on table "public"."usage_events" to "anon";

grant insert on table "public"."usage_events" to "anon";

grant references on table "public"."usage_events" to "anon";

grant select on table "public"."usage_events" to "anon";

grant trigger on table "public"."usage_events" to "anon";

grant truncate on table "public"."usage_events" to "anon";

grant update on table "public"."usage_events" to "anon";

grant delete on table "public"."usage_events" to "authenticated";

grant insert on table "public"."usage_events" to "authenticated";

grant references on table "public"."usage_events" to "authenticated";

grant select on table "public"."usage_events" to "authenticated";

grant trigger on table "public"."usage_events" to "authenticated";

grant truncate on table "public"."usage_events" to "authenticated";

grant update on table "public"."usage_events" to "authenticated";

grant delete on table "public"."usage_events" to "service_role";

grant insert on table "public"."usage_events" to "service_role";

grant references on table "public"."usage_events" to "service_role";

grant select on table "public"."usage_events" to "service_role";

grant trigger on table "public"."usage_events" to "service_role";

grant truncate on table "public"."usage_events" to "service_role";

grant update on table "public"."usage_events" to "service_role";

create policy "Deny all access to usage_events"
on "public"."usage_events"
as permissive
for all
to public
using (false)
with check (false);



