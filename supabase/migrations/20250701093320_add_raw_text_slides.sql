create extension if not exists "vector" with schema "public" version '0.8.0';

alter table "public"."slides" add column "raw_text" text;


