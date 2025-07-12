create type "public"."dlq_message_status" as enum ('unprocessed', 'processed', 'ignored');

create table "public"."dead_letter_messages" (
    "id" uuid not null default gen_random_uuid(),
    "subscription_name" text not null,
    "message_id" text not null,
    "payload" jsonb not null,
    "attributes" jsonb,
    "status" dlq_message_status not null default 'unprocessed'::dlq_message_status,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


alter table "public"."dead_letter_messages" enable row level security;

CREATE UNIQUE INDEX dead_letter_messages_pkey ON public.dead_letter_messages USING btree (id);

CREATE INDEX idx_dead_letter_messages_created_at ON public.dead_letter_messages USING btree (created_at);

CREATE INDEX idx_dead_letter_messages_status ON public.dead_letter_messages USING btree (status);

alter table "public"."dead_letter_messages" add constraint "dead_letter_messages_pkey" PRIMARY KEY using index "dead_letter_messages_pkey";

grant delete on table "public"."dead_letter_messages" to "anon";

grant insert on table "public"."dead_letter_messages" to "anon";

grant references on table "public"."dead_letter_messages" to "anon";

grant select on table "public"."dead_letter_messages" to "anon";

grant trigger on table "public"."dead_letter_messages" to "anon";

grant truncate on table "public"."dead_letter_messages" to "anon";

grant update on table "public"."dead_letter_messages" to "anon";

grant delete on table "public"."dead_letter_messages" to "authenticated";

grant insert on table "public"."dead_letter_messages" to "authenticated";

grant references on table "public"."dead_letter_messages" to "authenticated";

grant select on table "public"."dead_letter_messages" to "authenticated";

grant trigger on table "public"."dead_letter_messages" to "authenticated";

grant truncate on table "public"."dead_letter_messages" to "authenticated";

grant update on table "public"."dead_letter_messages" to "authenticated";

grant delete on table "public"."dead_letter_messages" to "service_role";

grant insert on table "public"."dead_letter_messages" to "service_role";

grant references on table "public"."dead_letter_messages" to "service_role";

grant select on table "public"."dead_letter_messages" to "service_role";

grant trigger on table "public"."dead_letter_messages" to "service_role";

grant truncate on table "public"."dead_letter_messages" to "service_role";

grant update on table "public"."dead_letter_messages" to "service_role";

create policy "Deny all access to dead_letter_messages"
on "public"."dead_letter_messages"
as permissive
for all
to public
using (false)
with check (false);



