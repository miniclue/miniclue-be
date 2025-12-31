alter table "public"."messages" drop constraint "messages_role_check";

alter table "public"."embeddings" alter column "vector" set data type public.vector(768) using "vector"::public.vector(768);

alter table "public"."messages" add constraint "messages_role_check" CHECK (((role)::text = ANY ((ARRAY['user'::character varying, 'assistant'::character varying])::text[]))) not valid;

alter table "public"."messages" validate constraint "messages_role_check";


