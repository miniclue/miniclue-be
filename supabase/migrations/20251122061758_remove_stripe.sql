drop policy "Deny all access to subscription_plans" on "public"."subscription_plans";

drop policy "Allow users to view own subscription" on "public"."user_subscriptions";

revoke delete on table "public"."subscription_plans" from "anon";

revoke insert on table "public"."subscription_plans" from "anon";

revoke references on table "public"."subscription_plans" from "anon";

revoke select on table "public"."subscription_plans" from "anon";

revoke trigger on table "public"."subscription_plans" from "anon";

revoke truncate on table "public"."subscription_plans" from "anon";

revoke update on table "public"."subscription_plans" from "anon";

revoke delete on table "public"."subscription_plans" from "authenticated";

revoke insert on table "public"."subscription_plans" from "authenticated";

revoke references on table "public"."subscription_plans" from "authenticated";

revoke select on table "public"."subscription_plans" from "authenticated";

revoke trigger on table "public"."subscription_plans" from "authenticated";

revoke truncate on table "public"."subscription_plans" from "authenticated";

revoke update on table "public"."subscription_plans" from "authenticated";

revoke delete on table "public"."subscription_plans" from "service_role";

revoke insert on table "public"."subscription_plans" from "service_role";

revoke references on table "public"."subscription_plans" from "service_role";

revoke select on table "public"."subscription_plans" from "service_role";

revoke trigger on table "public"."subscription_plans" from "service_role";

revoke truncate on table "public"."subscription_plans" from "service_role";

revoke update on table "public"."subscription_plans" from "service_role";

revoke delete on table "public"."user_subscriptions" from "anon";

revoke insert on table "public"."user_subscriptions" from "anon";

revoke references on table "public"."user_subscriptions" from "anon";

revoke select on table "public"."user_subscriptions" from "anon";

revoke trigger on table "public"."user_subscriptions" from "anon";

revoke truncate on table "public"."user_subscriptions" from "anon";

revoke update on table "public"."user_subscriptions" from "anon";

revoke delete on table "public"."user_subscriptions" from "authenticated";

revoke insert on table "public"."user_subscriptions" from "authenticated";

revoke references on table "public"."user_subscriptions" from "authenticated";

revoke select on table "public"."user_subscriptions" from "authenticated";

revoke trigger on table "public"."user_subscriptions" from "authenticated";

revoke truncate on table "public"."user_subscriptions" from "authenticated";

revoke update on table "public"."user_subscriptions" from "authenticated";

revoke delete on table "public"."user_subscriptions" from "service_role";

revoke insert on table "public"."user_subscriptions" from "service_role";

revoke references on table "public"."user_subscriptions" from "service_role";

revoke select on table "public"."user_subscriptions" from "service_role";

revoke trigger on table "public"."user_subscriptions" from "service_role";

revoke truncate on table "public"."user_subscriptions" from "service_role";

revoke update on table "public"."user_subscriptions" from "service_role";

alter table "public"."user_subscriptions" drop constraint "user_subscriptions_plan_id_fkey";

alter table "public"."user_subscriptions" drop constraint "user_subscriptions_user_id_fkey";

alter table "public"."subscription_plans" drop constraint "subscription_plans_pkey";

alter table "public"."user_subscriptions" drop constraint "user_subscriptions_pkey";

drop index if exists "public"."subscription_plans_pkey";

drop index if exists "public"."user_subscriptions_pkey";

drop table "public"."subscription_plans";

drop table "public"."user_subscriptions";

alter table "public"."user_profiles" drop column "stripe_customer_id";

drop type "public"."subscription_status";


