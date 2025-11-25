create extension if not exists "pg_cron" with schema "pg_catalog";

alter table "public"."user_profiles" add column "api_keys_provided" jsonb default '{}'::jsonb;


