drop policy "Deny all access to usage_events" on "public"."usage_events";

revoke delete on table "public"."usage_events" from "anon";

revoke insert on table "public"."usage_events" from "anon";

revoke references on table "public"."usage_events" from "anon";

revoke select on table "public"."usage_events" from "anon";

revoke trigger on table "public"."usage_events" from "anon";

revoke truncate on table "public"."usage_events" from "anon";

revoke update on table "public"."usage_events" from "anon";

revoke delete on table "public"."usage_events" from "authenticated";

revoke insert on table "public"."usage_events" from "authenticated";

revoke references on table "public"."usage_events" from "authenticated";

revoke select on table "public"."usage_events" from "authenticated";

revoke trigger on table "public"."usage_events" from "authenticated";

revoke truncate on table "public"."usage_events" from "authenticated";

revoke update on table "public"."usage_events" from "authenticated";

revoke delete on table "public"."usage_events" from "service_role";

revoke insert on table "public"."usage_events" from "service_role";

revoke references on table "public"."usage_events" from "service_role";

revoke select on table "public"."usage_events" from "service_role";

revoke trigger on table "public"."usage_events" from "service_role";

revoke truncate on table "public"."usage_events" from "service_role";

revoke update on table "public"."usage_events" from "service_role";

alter table "public"."usage_events" drop constraint "usage_events_lecture_id_fkey";

alter table "public"."usage_events" drop constraint "usage_events_user_id_fkey";

alter table "public"."messages" drop constraint "messages_role_check";

alter table "public"."usage_events" drop constraint "usage_events_pkey";

drop index if exists "public"."idx_usage_events_user_event_time";

drop index if exists "public"."usage_events_pkey";

drop table "public"."usage_events";

drop type "public"."usage_event_type";

alter table "public"."messages" add constraint "messages_role_check" CHECK (((role)::text = ANY ((ARRAY['user'::character varying, 'assistant'::character varying])::text[]))) not valid;

alter table "public"."messages" validate constraint "messages_role_check";


