create extension if not exists "vector" with schema "public" version '0.8.0';

alter table "public"."lectures" drop column "error_message";

alter table "public"."lectures" add column "error_details" jsonb;


