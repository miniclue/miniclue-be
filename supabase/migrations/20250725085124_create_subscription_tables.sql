create type "public"."subscription_status" as enum ('active', 'cancelled', 'past_due');

create table "public"."subscription_plans" (
    "id" text not null,
    "name" text not null,
    "price_cents" integer not null,
    "billing_period" interval not null,
    "max_uploads" integer not null,
    "max_size_mb" integer not null,
    "chat_limit" integer not null default '-1'::integer,
    "feature_flags" jsonb not null default '{}'::jsonb
);


alter table "public"."subscription_plans" enable row level security;

create table "public"."user_subscriptions" (
    "user_id" uuid not null,
    "plan_id" text not null,
    "starts_at" timestamp with time zone not null default now(),
    "ends_at" timestamp with time zone not null,
    "status" subscription_status not null
);


alter table "public"."user_subscriptions" enable row level security;

CREATE UNIQUE INDEX subscription_plans_pkey ON public.subscription_plans USING btree (id);

CREATE UNIQUE INDEX user_subscriptions_pkey ON public.user_subscriptions USING btree (user_id);

alter table "public"."subscription_plans" add constraint "subscription_plans_pkey" PRIMARY KEY using index "subscription_plans_pkey";

alter table "public"."user_subscriptions" add constraint "user_subscriptions_pkey" PRIMARY KEY using index "user_subscriptions_pkey";

alter table "public"."user_subscriptions" add constraint "user_subscriptions_plan_id_fkey" FOREIGN KEY (plan_id) REFERENCES subscription_plans(id) not valid;

alter table "public"."user_subscriptions" validate constraint "user_subscriptions_plan_id_fkey";

alter table "public"."user_subscriptions" add constraint "user_subscriptions_user_id_fkey" FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE not valid;

alter table "public"."user_subscriptions" validate constraint "user_subscriptions_user_id_fkey";

grant delete on table "public"."subscription_plans" to "anon";

grant insert on table "public"."subscription_plans" to "anon";

grant references on table "public"."subscription_plans" to "anon";

grant select on table "public"."subscription_plans" to "anon";

grant trigger on table "public"."subscription_plans" to "anon";

grant truncate on table "public"."subscription_plans" to "anon";

grant update on table "public"."subscription_plans" to "anon";

grant delete on table "public"."subscription_plans" to "authenticated";

grant insert on table "public"."subscription_plans" to "authenticated";

grant references on table "public"."subscription_plans" to "authenticated";

grant select on table "public"."subscription_plans" to "authenticated";

grant trigger on table "public"."subscription_plans" to "authenticated";

grant truncate on table "public"."subscription_plans" to "authenticated";

grant update on table "public"."subscription_plans" to "authenticated";

grant delete on table "public"."subscription_plans" to "service_role";

grant insert on table "public"."subscription_plans" to "service_role";

grant references on table "public"."subscription_plans" to "service_role";

grant select on table "public"."subscription_plans" to "service_role";

grant trigger on table "public"."subscription_plans" to "service_role";

grant truncate on table "public"."subscription_plans" to "service_role";

grant update on table "public"."subscription_plans" to "service_role";

grant delete on table "public"."user_subscriptions" to "anon";

grant insert on table "public"."user_subscriptions" to "anon";

grant references on table "public"."user_subscriptions" to "anon";

grant select on table "public"."user_subscriptions" to "anon";

grant trigger on table "public"."user_subscriptions" to "anon";

grant truncate on table "public"."user_subscriptions" to "anon";

grant update on table "public"."user_subscriptions" to "anon";

grant delete on table "public"."user_subscriptions" to "authenticated";

grant insert on table "public"."user_subscriptions" to "authenticated";

grant references on table "public"."user_subscriptions" to "authenticated";

grant select on table "public"."user_subscriptions" to "authenticated";

grant trigger on table "public"."user_subscriptions" to "authenticated";

grant truncate on table "public"."user_subscriptions" to "authenticated";

grant update on table "public"."user_subscriptions" to "authenticated";

grant delete on table "public"."user_subscriptions" to "service_role";

grant insert on table "public"."user_subscriptions" to "service_role";

grant references on table "public"."user_subscriptions" to "service_role";

grant select on table "public"."user_subscriptions" to "service_role";

grant trigger on table "public"."user_subscriptions" to "service_role";

grant truncate on table "public"."user_subscriptions" to "service_role";

grant update on table "public"."user_subscriptions" to "service_role";

create policy "Deny all access to subscription_plans"
on "public"."subscription_plans"
as permissive
for all
to public
using (false)
with check (false);


create policy "Allow users to view own subscription"
on "public"."user_subscriptions"
as permissive
for select
to public
using ((auth.uid() = user_id));

-- Seed initial subscription plans
INSERT INTO public.subscription_plans (id, name, price_cents, billing_period, max_uploads, max_size_mb, chat_limit, feature_flags) VALUES
  ('free', 'Free', 0, '31 days', 3, 10, -1, '{}'::jsonb),
  ('beta', 'Beta', 0, '31 days', 100, 300, -1, '{}'::jsonb),
  ('pro_monthly_early', 'Pro (Monthly) Early Adopter', 1000, '31 days', -1, 300, -1, '{}'::jsonb),
  ('pro_annual_early', 'Pro (Annual) Early Adopter', 600, '365 days', -1, 300, -1, '{}'::jsonb),
  ('pro_monthly', 'Pro (Monthly)', 2000, '31 days', -1, 300, -1, '{}'::jsonb),
  ('pro_annual', 'Pro (Annual)', 1200, '365 days', -1, 300, -1, '{}'::jsonb)
ON CONFLICT (id) DO NOTHING;

-- Assign Beta plan to all existing users for a 31-day period
INSERT INTO public.user_subscriptions (user_id, plan_id, starts_at, ends_at, status)
SELECT id, 'beta', NOW(), NOW() + INTERVAL '31 days', 'active'
FROM auth.users
ON CONFLICT (user_id) DO UPDATE
  SET plan_id = 'beta', 
      starts_at = EXCLUDED.starts_at, 
      ends_at = EXCLUDED.ends_at, 
      status = EXCLUDED.status;