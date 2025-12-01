drop policy "Deny all access to llm_calls" on "public"."llm_calls";

revoke delete on table "public"."llm_calls" from "anon";

revoke insert on table "public"."llm_calls" from "anon";

revoke references on table "public"."llm_calls" from "anon";

revoke select on table "public"."llm_calls" from "anon";

revoke trigger on table "public"."llm_calls" from "anon";

revoke truncate on table "public"."llm_calls" from "anon";

revoke update on table "public"."llm_calls" from "anon";

revoke delete on table "public"."llm_calls" from "authenticated";

revoke insert on table "public"."llm_calls" from "authenticated";

revoke references on table "public"."llm_calls" from "authenticated";

revoke select on table "public"."llm_calls" from "authenticated";

revoke trigger on table "public"."llm_calls" from "authenticated";

revoke truncate on table "public"."llm_calls" from "authenticated";

revoke update on table "public"."llm_calls" from "authenticated";

revoke delete on table "public"."llm_calls" from "service_role";

revoke insert on table "public"."llm_calls" from "service_role";

revoke references on table "public"."llm_calls" from "service_role";

revoke select on table "public"."llm_calls" from "service_role";

revoke trigger on table "public"."llm_calls" from "service_role";

revoke truncate on table "public"."llm_calls" from "service_role";

revoke update on table "public"."llm_calls" from "service_role";

alter table "public"."llm_calls" drop constraint "llm_calls_call_type_check";

alter table "public"."llm_calls" drop constraint "llm_calls_lecture_id_fkey";

alter table "public"."llm_calls" drop constraint "llm_calls_slide_id_fkey";

alter table "public"."messages" drop constraint "messages_role_check";

alter table "public"."llm_calls" drop constraint "llm_calls_pkey";

drop index if exists "public"."idx_llm_calls_occurred_at";

drop index if exists "public"."llm_calls_pkey";

drop table "public"."llm_calls";

alter table "public"."messages" add constraint "messages_role_check" CHECK (((role)::text = ANY ((ARRAY['user'::character varying, 'assistant'::character varying])::text[]))) not valid;

alter table "public"."messages" validate constraint "messages_role_check";


