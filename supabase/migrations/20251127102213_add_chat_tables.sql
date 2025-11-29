
  create table "public"."chats" (
    "id" uuid not null default gen_random_uuid(),
    "lecture_id" uuid not null,
    "user_id" uuid not null,
    "title" text not null,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
      );


alter table "public"."chats" enable row level security;


  create table "public"."messages" (
    "id" uuid not null default gen_random_uuid(),
    "chat_id" uuid not null,
    "role" character varying not null,
    "parts" jsonb not null,
    "created_at" timestamp with time zone not null default now()
      );


alter table "public"."messages" enable row level security;

CREATE UNIQUE INDEX chats_pkey ON public.chats USING btree (id);

CREATE INDEX idx_chats_lecture_id ON public.chats USING btree (lecture_id);

CREATE INDEX idx_chats_lecture_user ON public.chats USING btree (lecture_id, user_id);

CREATE INDEX idx_chats_user_id ON public.chats USING btree (user_id);

CREATE INDEX idx_messages_chat_id ON public.messages USING btree (chat_id);

CREATE INDEX idx_messages_created_at ON public.messages USING btree (created_at);

CREATE UNIQUE INDEX messages_pkey ON public.messages USING btree (id);

alter table "public"."chats" add constraint "chats_pkey" PRIMARY KEY using index "chats_pkey";

alter table "public"."messages" add constraint "messages_pkey" PRIMARY KEY using index "messages_pkey";

alter table "public"."chats" add constraint "chats_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES public.lectures(id) ON DELETE CASCADE not valid;

alter table "public"."chats" validate constraint "chats_lecture_id_fkey";

alter table "public"."chats" add constraint "chats_user_id_fkey" FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE not valid;

alter table "public"."chats" validate constraint "chats_user_id_fkey";

alter table "public"."messages" add constraint "messages_chat_id_fkey" FOREIGN KEY (chat_id) REFERENCES public.chats(id) ON DELETE CASCADE not valid;

alter table "public"."messages" validate constraint "messages_chat_id_fkey";

alter table "public"."messages" add constraint "messages_role_check" CHECK (((role)::text = ANY ((ARRAY['user'::character varying, 'assistant'::character varying])::text[]))) not valid;

alter table "public"."messages" validate constraint "messages_role_check";

grant delete on table "public"."chats" to "anon";

grant insert on table "public"."chats" to "anon";

grant references on table "public"."chats" to "anon";

grant select on table "public"."chats" to "anon";

grant trigger on table "public"."chats" to "anon";

grant truncate on table "public"."chats" to "anon";

grant update on table "public"."chats" to "anon";

grant delete on table "public"."chats" to "authenticated";

grant insert on table "public"."chats" to "authenticated";

grant references on table "public"."chats" to "authenticated";

grant select on table "public"."chats" to "authenticated";

grant trigger on table "public"."chats" to "authenticated";

grant truncate on table "public"."chats" to "authenticated";

grant update on table "public"."chats" to "authenticated";

grant delete on table "public"."chats" to "service_role";

grant insert on table "public"."chats" to "service_role";

grant references on table "public"."chats" to "service_role";

grant select on table "public"."chats" to "service_role";

grant trigger on table "public"."chats" to "service_role";

grant truncate on table "public"."chats" to "service_role";

grant update on table "public"."chats" to "service_role";

grant delete on table "public"."messages" to "anon";

grant insert on table "public"."messages" to "anon";

grant references on table "public"."messages" to "anon";

grant select on table "public"."messages" to "anon";

grant trigger on table "public"."messages" to "anon";

grant truncate on table "public"."messages" to "anon";

grant update on table "public"."messages" to "anon";

grant delete on table "public"."messages" to "authenticated";

grant insert on table "public"."messages" to "authenticated";

grant references on table "public"."messages" to "authenticated";

grant select on table "public"."messages" to "authenticated";

grant trigger on table "public"."messages" to "authenticated";

grant truncate on table "public"."messages" to "authenticated";

grant update on table "public"."messages" to "authenticated";

grant delete on table "public"."messages" to "service_role";

grant insert on table "public"."messages" to "service_role";

grant references on table "public"."messages" to "service_role";

grant select on table "public"."messages" to "service_role";

grant trigger on table "public"."messages" to "service_role";

grant truncate on table "public"."messages" to "service_role";

grant update on table "public"."messages" to "service_role";


  create policy "Allow all access to own chats"
  on "public"."chats"
  as permissive
  for all
  to public
using (((auth.uid() = user_id) AND (EXISTS ( SELECT 1
   FROM public.lectures
  WHERE ((lectures.id = chats.lecture_id) AND (lectures.user_id = auth.uid()))))))
with check (((auth.uid() = user_id) AND (EXISTS ( SELECT 1
   FROM public.lectures
  WHERE ((lectures.id = chats.lecture_id) AND (lectures.user_id = auth.uid()))))));



  create policy "Allow all access to own messages"
  on "public"."messages"
  as permissive
  for all
  to public
using ((EXISTS ( SELECT 1
   FROM public.chats
  WHERE ((chats.id = messages.chat_id) AND (chats.user_id = auth.uid()) AND (EXISTS ( SELECT 1
           FROM public.lectures
          WHERE ((lectures.id = chats.lecture_id) AND (lectures.user_id = auth.uid()))))))))
with check ((EXISTS ( SELECT 1
   FROM public.chats
  WHERE ((chats.id = messages.chat_id) AND (chats.user_id = auth.uid()) AND (EXISTS ( SELECT 1
           FROM public.lectures
          WHERE ((lectures.id = chats.lecture_id) AND (lectures.user_id = auth.uid()))))))));



