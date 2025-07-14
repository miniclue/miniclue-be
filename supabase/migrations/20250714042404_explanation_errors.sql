alter table "public"."lectures" drop column "error_details";

alter table "public"."lectures" add column "explanation_error_details" jsonb;

alter table "public"."lectures" add column "search_error_details" jsonb;


