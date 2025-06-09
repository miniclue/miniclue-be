create extension if not exists "vector" with schema "public" version '0.8.0';

alter table "public"."lectures" add column "accessed_at" timestamp with time zone not null default now();


