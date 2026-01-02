drop policy "Allow read access to explanations of own lectures" on "public"."explanations";

drop policy "Allow read access to summaries of own lectures" on "public"."summaries";

revoke delete on table "public"."explanations" from "anon";

revoke insert on table "public"."explanations" from "anon";

revoke references on table "public"."explanations" from "anon";

revoke select on table "public"."explanations" from "anon";

revoke trigger on table "public"."explanations" from "anon";

revoke truncate on table "public"."explanations" from "anon";

revoke update on table "public"."explanations" from "anon";

revoke delete on table "public"."explanations" from "authenticated";

revoke insert on table "public"."explanations" from "authenticated";

revoke references on table "public"."explanations" from "authenticated";

revoke select on table "public"."explanations" from "authenticated";

revoke trigger on table "public"."explanations" from "authenticated";

revoke truncate on table "public"."explanations" from "authenticated";

revoke update on table "public"."explanations" from "authenticated";

revoke delete on table "public"."explanations" from "service_role";

revoke insert on table "public"."explanations" from "service_role";

revoke references on table "public"."explanations" from "service_role";

revoke select on table "public"."explanations" from "service_role";

revoke trigger on table "public"."explanations" from "service_role";

revoke truncate on table "public"."explanations" from "service_role";

revoke update on table "public"."explanations" from "service_role";

revoke delete on table "public"."summaries" from "anon";

revoke insert on table "public"."summaries" from "anon";

revoke references on table "public"."summaries" from "anon";

revoke select on table "public"."summaries" from "anon";

revoke trigger on table "public"."summaries" from "anon";

revoke truncate on table "public"."summaries" from "anon";

revoke update on table "public"."summaries" from "anon";

revoke delete on table "public"."summaries" from "authenticated";

revoke insert on table "public"."summaries" from "authenticated";

revoke references on table "public"."summaries" from "authenticated";

revoke select on table "public"."summaries" from "authenticated";

revoke trigger on table "public"."summaries" from "authenticated";

revoke truncate on table "public"."summaries" from "authenticated";

revoke update on table "public"."summaries" from "authenticated";

revoke delete on table "public"."summaries" from "service_role";

revoke insert on table "public"."summaries" from "service_role";

revoke references on table "public"."summaries" from "service_role";

revoke select on table "public"."summaries" from "service_role";

revoke trigger on table "public"."summaries" from "service_role";

revoke truncate on table "public"."summaries" from "service_role";

revoke update on table "public"."summaries" from "service_role";

alter table "public"."explanations" drop constraint "explanations_lecture_id_slide_number_fkey";

alter table "public"."explanations" drop constraint "explanations_slide_id_fkey";

alter table "public"."explanations" drop constraint "explanations_slide_id_key";

alter table "public"."summaries" drop constraint "summaries_lecture_id_fkey";

alter table "public"."messages" drop constraint "messages_role_check";

alter table "public"."explanations" drop constraint "explanations_pkey";

alter table "public"."summaries" drop constraint "summaries_pkey";

drop index if exists "public"."explanations_pkey";

drop index if exists "public"."explanations_slide_id_key";

drop index if exists "public"."idx_explanations_lecture_slide";

drop index if exists "public"."idx_summaries_by_lecture";

drop index if exists "public"."summaries_pkey";

drop table "public"."explanations";

drop table "public"."summaries";

alter table "public"."lectures" alter column "status" drop default;

alter type "public"."lecture_status" rename to "lecture_status__old_version_to_be_dropped";

create type "public"."lecture_status" as enum ('uploading', 'pending_processing', 'parsing', 'processing', 'complete', 'failed');

alter table "public"."lectures" alter column status type "public"."lecture_status" using status::text::"public"."lecture_status";

alter table "public"."lectures" alter column "status" set default 'uploading'::public.lecture_status;

drop type "public"."lecture_status__old_version_to_be_dropped";

alter table "public"."lectures" drop column "explanation_error_details";

alter table "public"."lectures" drop column "processed_slides";

alter table "public"."messages" add constraint "messages_role_check" CHECK (((role)::text = ANY ((ARRAY['user'::character varying, 'assistant'::character varying])::text[]))) not valid;

alter table "public"."messages" validate constraint "messages_role_check";


