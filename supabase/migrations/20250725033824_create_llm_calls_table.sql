create table "public"."llm_calls" (
    "id" uuid not null default gen_random_uuid(),
    "lecture_id" uuid,
    "slide_id" uuid,
    "call_type" text not null,
    "model" text not null,
    "prompt_tokens" integer not null,
    "completion_tokens" integer not null,
    "total_tokens" integer not null,
    "currency" text not null default 'USD'::text,
    "cost" numeric(10,6) not null,
    "occurred_at" timestamp with time zone not null default now(),
    "metadata" jsonb not null default '{}'::jsonb
);


alter table "public"."llm_calls" enable row level security;

alter table "public"."slide_images" add column "metadata" jsonb not null default '{}'::jsonb;

CREATE INDEX idx_llm_calls_occurred_at ON public.llm_calls USING btree (occurred_at);

CREATE UNIQUE INDEX llm_calls_pkey ON public.llm_calls USING btree (id);

alter table "public"."llm_calls" add constraint "llm_calls_pkey" PRIMARY KEY using index "llm_calls_pkey";

alter table "public"."llm_calls" add constraint "llm_calls_call_type_check" CHECK ((call_type = ANY (ARRAY['ingestion'::text, 'explanation'::text, 'embedding'::text, 'summary'::text, 'image_analysis'::text, 'other'::text]))) not valid;

alter table "public"."llm_calls" validate constraint "llm_calls_call_type_check";

alter table "public"."llm_calls" add constraint "llm_calls_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES lectures(id) ON DELETE SET NULL not valid;

alter table "public"."llm_calls" validate constraint "llm_calls_lecture_id_fkey";

alter table "public"."llm_calls" add constraint "llm_calls_slide_id_fkey" FOREIGN KEY (slide_id) REFERENCES slides(id) ON DELETE SET NULL not valid;

alter table "public"."llm_calls" validate constraint "llm_calls_slide_id_fkey";

grant delete on table "public"."llm_calls" to "anon";

grant insert on table "public"."llm_calls" to "anon";

grant references on table "public"."llm_calls" to "anon";

grant select on table "public"."llm_calls" to "anon";

grant trigger on table "public"."llm_calls" to "anon";

grant truncate on table "public"."llm_calls" to "anon";

grant update on table "public"."llm_calls" to "anon";

grant delete on table "public"."llm_calls" to "authenticated";

grant insert on table "public"."llm_calls" to "authenticated";

grant references on table "public"."llm_calls" to "authenticated";

grant select on table "public"."llm_calls" to "authenticated";

grant trigger on table "public"."llm_calls" to "authenticated";

grant truncate on table "public"."llm_calls" to "authenticated";

grant update on table "public"."llm_calls" to "authenticated";

grant delete on table "public"."llm_calls" to "service_role";

grant insert on table "public"."llm_calls" to "service_role";

grant references on table "public"."llm_calls" to "service_role";

grant select on table "public"."llm_calls" to "service_role";

grant trigger on table "public"."llm_calls" to "service_role";

grant truncate on table "public"."llm_calls" to "service_role";

grant update on table "public"."llm_calls" to "service_role";

create policy "Deny all access to llm_calls"
on "public"."llm_calls"
as permissive
for all
to public
using (false)
with check (false);



