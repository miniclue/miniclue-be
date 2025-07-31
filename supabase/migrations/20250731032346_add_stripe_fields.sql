alter table "public"."user_profiles" add column "stripe_customer_id" text;

alter table "public"."user_subscriptions" add column "stripe_subscription_id" text;


