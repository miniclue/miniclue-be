create extension if not exists "vector" with schema "public" version '0.8.0';

alter table "public"."user_profiles" drop column "full_name";

alter table "public"."user_profiles" add column "email" text default ''::text;

alter table "public"."user_profiles" add column "name" text default ''::text;


