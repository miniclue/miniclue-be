create extension if not exists "vector" with schema "extensions";


create table "public"."explanations" (
    "id" uuid not null default gen_random_uuid(),
    "lecture_id" uuid not null,
    "slide_number" integer not null,
    "content" text not null,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


create table "public"."lectures" (
    "id" uuid not null default gen_random_uuid(),
    "user_id" uuid not null,
    "title" text not null,
    "status" character varying(32) not null default 'uploaded'::character varying,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


create table "public"."notes" (
    "id" uuid not null default gen_random_uuid(),
    "user_id" uuid not null,
    "lecture_id" uuid not null,
    "content" text not null,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


create table "public"."slide_images" (
    "id" uuid not null default gen_random_uuid(),
    "lecture_id" uuid not null,
    "slide_number" integer not null,
    "image_index" integer not null,
    "storage_path" text not null,
    "caption" text not null default ''::text,
    "width" integer not null,
    "height" integer not null,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


create table "public"."slides" (
    "id" uuid not null default gen_random_uuid(),
    "lecture_id" uuid not null,
    "slide_number" integer not null,
    "image_keys" text[] not null default '{}'::text[],
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


create table "public"."summaries" (
    "lecture_id" uuid not null,
    "content" text not null,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


CREATE UNIQUE INDEX explanations_lecture_id_slide_number_key ON public.explanations USING btree (lecture_id, slide_number);

CREATE UNIQUE INDEX explanations_pkey ON public.explanations USING btree (id);

CREATE INDEX idx_explanations_lecture_slide ON public.explanations USING btree (lecture_id, slide_number);

CREATE INDEX idx_lectures_user_id ON public.lectures USING btree (user_id);

CREATE INDEX idx_notes_lecture_id ON public.notes USING btree (lecture_id);

CREATE INDEX idx_notes_user_id ON public.notes USING btree (user_id);

CREATE INDEX idx_slide_images_lecture_slide ON public.slide_images USING btree (lecture_id, slide_number);

CREATE INDEX idx_slides_lecture_id ON public.slides USING btree (lecture_id);

CREATE INDEX idx_slides_lecture_slide ON public.slides USING btree (lecture_id, slide_number);

CREATE INDEX idx_summaries_by_lecture ON public.summaries USING btree (lecture_id);

CREATE UNIQUE INDEX lectures_pkey ON public.lectures USING btree (id);

CREATE UNIQUE INDEX notes_pkey ON public.notes USING btree (id);

CREATE UNIQUE INDEX slide_images_lecture_id_slide_number_image_index_key ON public.slide_images USING btree (lecture_id, slide_number, image_index);

CREATE UNIQUE INDEX slide_images_pkey ON public.slide_images USING btree (id);

CREATE UNIQUE INDEX slides_lecture_id_slide_number_key ON public.slides USING btree (lecture_id, slide_number);

CREATE UNIQUE INDEX slides_pkey ON public.slides USING btree (id);

CREATE UNIQUE INDEX summaries_pkey ON public.summaries USING btree (lecture_id);

alter table "public"."explanations" add constraint "explanations_pkey" PRIMARY KEY using index "explanations_pkey";

alter table "public"."lectures" add constraint "lectures_pkey" PRIMARY KEY using index "lectures_pkey";

alter table "public"."notes" add constraint "notes_pkey" PRIMARY KEY using index "notes_pkey";

alter table "public"."slide_images" add constraint "slide_images_pkey" PRIMARY KEY using index "slide_images_pkey";

alter table "public"."slides" add constraint "slides_pkey" PRIMARY KEY using index "slides_pkey";

alter table "public"."summaries" add constraint "summaries_pkey" PRIMARY KEY using index "summaries_pkey";

alter table "public"."explanations" add constraint "explanations_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES lectures(id) ON DELETE CASCADE not valid;

alter table "public"."explanations" validate constraint "explanations_lecture_id_fkey";

alter table "public"."explanations" add constraint "explanations_lecture_id_slide_number_key" UNIQUE using index "explanations_lecture_id_slide_number_key";

alter table "public"."explanations" add constraint "fk_explanations_slide" FOREIGN KEY (lecture_id, slide_number) REFERENCES slides(lecture_id, slide_number) ON DELETE CASCADE not valid;

alter table "public"."explanations" validate constraint "fk_explanations_slide";

alter table "public"."notes" add constraint "notes_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES lectures(id) ON DELETE CASCADE not valid;

alter table "public"."notes" validate constraint "notes_lecture_id_fkey";

alter table "public"."slide_images" add constraint "fk_slide_images_slide" FOREIGN KEY (lecture_id, slide_number) REFERENCES slides(lecture_id, slide_number) ON DELETE CASCADE not valid;

alter table "public"."slide_images" validate constraint "fk_slide_images_slide";

alter table "public"."slide_images" add constraint "slide_images_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES lectures(id) ON DELETE CASCADE not valid;

alter table "public"."slide_images" validate constraint "slide_images_lecture_id_fkey";

alter table "public"."slide_images" add constraint "slide_images_lecture_id_slide_number_image_index_key" UNIQUE using index "slide_images_lecture_id_slide_number_image_index_key";

alter table "public"."slides" add constraint "slides_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES lectures(id) ON DELETE CASCADE not valid;

alter table "public"."slides" validate constraint "slides_lecture_id_fkey";

alter table "public"."slides" add constraint "slides_lecture_id_slide_number_key" UNIQUE using index "slides_lecture_id_slide_number_key";

alter table "public"."summaries" add constraint "summaries_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES lectures(id) ON DELETE CASCADE not valid;

alter table "public"."summaries" validate constraint "summaries_lecture_id_fkey";

grant delete on table "public"."explanations" to "anon";

grant insert on table "public"."explanations" to "anon";

grant references on table "public"."explanations" to "anon";

grant select on table "public"."explanations" to "anon";

grant trigger on table "public"."explanations" to "anon";

grant truncate on table "public"."explanations" to "anon";

grant update on table "public"."explanations" to "anon";

grant delete on table "public"."explanations" to "authenticated";

grant insert on table "public"."explanations" to "authenticated";

grant references on table "public"."explanations" to "authenticated";

grant select on table "public"."explanations" to "authenticated";

grant trigger on table "public"."explanations" to "authenticated";

grant truncate on table "public"."explanations" to "authenticated";

grant update on table "public"."explanations" to "authenticated";

grant delete on table "public"."explanations" to "service_role";

grant insert on table "public"."explanations" to "service_role";

grant references on table "public"."explanations" to "service_role";

grant select on table "public"."explanations" to "service_role";

grant trigger on table "public"."explanations" to "service_role";

grant truncate on table "public"."explanations" to "service_role";

grant update on table "public"."explanations" to "service_role";

grant delete on table "public"."lectures" to "anon";

grant insert on table "public"."lectures" to "anon";

grant references on table "public"."lectures" to "anon";

grant select on table "public"."lectures" to "anon";

grant trigger on table "public"."lectures" to "anon";

grant truncate on table "public"."lectures" to "anon";

grant update on table "public"."lectures" to "anon";

grant delete on table "public"."lectures" to "authenticated";

grant insert on table "public"."lectures" to "authenticated";

grant references on table "public"."lectures" to "authenticated";

grant select on table "public"."lectures" to "authenticated";

grant trigger on table "public"."lectures" to "authenticated";

grant truncate on table "public"."lectures" to "authenticated";

grant update on table "public"."lectures" to "authenticated";

grant delete on table "public"."lectures" to "service_role";

grant insert on table "public"."lectures" to "service_role";

grant references on table "public"."lectures" to "service_role";

grant select on table "public"."lectures" to "service_role";

grant trigger on table "public"."lectures" to "service_role";

grant truncate on table "public"."lectures" to "service_role";

grant update on table "public"."lectures" to "service_role";

grant delete on table "public"."notes" to "anon";

grant insert on table "public"."notes" to "anon";

grant references on table "public"."notes" to "anon";

grant select on table "public"."notes" to "anon";

grant trigger on table "public"."notes" to "anon";

grant truncate on table "public"."notes" to "anon";

grant update on table "public"."notes" to "anon";

grant delete on table "public"."notes" to "authenticated";

grant insert on table "public"."notes" to "authenticated";

grant references on table "public"."notes" to "authenticated";

grant select on table "public"."notes" to "authenticated";

grant trigger on table "public"."notes" to "authenticated";

grant truncate on table "public"."notes" to "authenticated";

grant update on table "public"."notes" to "authenticated";

grant delete on table "public"."notes" to "service_role";

grant insert on table "public"."notes" to "service_role";

grant references on table "public"."notes" to "service_role";

grant select on table "public"."notes" to "service_role";

grant trigger on table "public"."notes" to "service_role";

grant truncate on table "public"."notes" to "service_role";

grant update on table "public"."notes" to "service_role";

grant delete on table "public"."slide_images" to "anon";

grant insert on table "public"."slide_images" to "anon";

grant references on table "public"."slide_images" to "anon";

grant select on table "public"."slide_images" to "anon";

grant trigger on table "public"."slide_images" to "anon";

grant truncate on table "public"."slide_images" to "anon";

grant update on table "public"."slide_images" to "anon";

grant delete on table "public"."slide_images" to "authenticated";

grant insert on table "public"."slide_images" to "authenticated";

grant references on table "public"."slide_images" to "authenticated";

grant select on table "public"."slide_images" to "authenticated";

grant trigger on table "public"."slide_images" to "authenticated";

grant truncate on table "public"."slide_images" to "authenticated";

grant update on table "public"."slide_images" to "authenticated";

grant delete on table "public"."slide_images" to "service_role";

grant insert on table "public"."slide_images" to "service_role";

grant references on table "public"."slide_images" to "service_role";

grant select on table "public"."slide_images" to "service_role";

grant trigger on table "public"."slide_images" to "service_role";

grant truncate on table "public"."slide_images" to "service_role";

grant update on table "public"."slide_images" to "service_role";

grant delete on table "public"."slides" to "anon";

grant insert on table "public"."slides" to "anon";

grant references on table "public"."slides" to "anon";

grant select on table "public"."slides" to "anon";

grant trigger on table "public"."slides" to "anon";

grant truncate on table "public"."slides" to "anon";

grant update on table "public"."slides" to "anon";

grant delete on table "public"."slides" to "authenticated";

grant insert on table "public"."slides" to "authenticated";

grant references on table "public"."slides" to "authenticated";

grant select on table "public"."slides" to "authenticated";

grant trigger on table "public"."slides" to "authenticated";

grant truncate on table "public"."slides" to "authenticated";

grant update on table "public"."slides" to "authenticated";

grant delete on table "public"."slides" to "service_role";

grant insert on table "public"."slides" to "service_role";

grant references on table "public"."slides" to "service_role";

grant select on table "public"."slides" to "service_role";

grant trigger on table "public"."slides" to "service_role";

grant truncate on table "public"."slides" to "service_role";

grant update on table "public"."slides" to "service_role";

grant delete on table "public"."summaries" to "anon";

grant insert on table "public"."summaries" to "anon";

grant references on table "public"."summaries" to "anon";

grant select on table "public"."summaries" to "anon";

grant trigger on table "public"."summaries" to "anon";

grant truncate on table "public"."summaries" to "anon";

grant update on table "public"."summaries" to "anon";

grant delete on table "public"."summaries" to "authenticated";

grant insert on table "public"."summaries" to "authenticated";

grant references on table "public"."summaries" to "authenticated";

grant select on table "public"."summaries" to "authenticated";

grant trigger on table "public"."summaries" to "authenticated";

grant truncate on table "public"."summaries" to "authenticated";

grant update on table "public"."summaries" to "authenticated";

grant delete on table "public"."summaries" to "service_role";

grant insert on table "public"."summaries" to "service_role";

grant references on table "public"."summaries" to "service_role";

grant select on table "public"."summaries" to "service_role";

grant trigger on table "public"."summaries" to "service_role";

grant truncate on table "public"."summaries" to "service_role";

grant update on table "public"."summaries" to "service_role";


