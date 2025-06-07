create extension if not exists "vector" with schema "public" version '0.8.0';

alter table "public"."lectures" add column "pdf_url" text not null;


