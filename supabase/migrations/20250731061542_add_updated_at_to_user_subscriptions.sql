alter table "public"."user_subscriptions" add column "created_at" timestamp with time zone not null default now();

alter table "public"."user_subscriptions" add column "updated_at" timestamp with time zone not null default now();


