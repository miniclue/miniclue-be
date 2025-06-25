create extension if not exists "vector" with schema "public" version '0.8.0';

alter table "public"."explanations" add column "one_liner" text not null default ''::text;

alter table "public"."slides" add column "pending_chunks_count" integer not null default 0;


