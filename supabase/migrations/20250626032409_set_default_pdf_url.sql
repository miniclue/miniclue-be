create extension if not exists "vector" with schema "public" version '0.8.0';

alter table "public"."lectures" alter column "pdf_url" set default ''::text;


