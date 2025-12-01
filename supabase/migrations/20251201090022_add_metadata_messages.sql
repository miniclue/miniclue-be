alter table "public"."messages" drop constraint "messages_role_check";

alter table "public"."messages" add column "metadata" jsonb not null default '{}'::jsonb;

alter table "public"."messages" add constraint "messages_role_check" CHECK (((role)::text = ANY ((ARRAY['user'::character varying, 'assistant'::character varying])::text[]))) not valid;

alter table "public"."messages" validate constraint "messages_role_check";


