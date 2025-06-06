create extension if not exists "vector" with schema "public" version '0.8.0';

create table "public"."chunks" (
    "id" uuid not null default gen_random_uuid(),
    "lecture_id" uuid not null,
    "slide_number" integer not null,
    "chunk_index" integer not null,
    "text" text not null,
    "is_image" boolean not null default false,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


create table "public"."embeddings" (
    "chunk_id" uuid not null,
    "lecture_id" uuid not null,
    "slide_number" integer not null,
    "vector" vector(1536) not null,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


CREATE UNIQUE INDEX chunks_lecture_id_slide_number_chunk_index_key ON public.chunks USING btree (lecture_id, slide_number, chunk_index);

CREATE UNIQUE INDEX chunks_pkey ON public.chunks USING btree (id);

CREATE UNIQUE INDEX embeddings_pkey ON public.embeddings USING btree (chunk_id);

CREATE INDEX idx_chunks_lecture ON public.chunks USING btree (lecture_id);

CREATE INDEX idx_chunks_lecture_slide ON public.chunks USING btree (lecture_id, slide_number);

CREATE INDEX idx_embeddings_lecture_snap ON public.embeddings USING btree (lecture_id, slide_number);

CREATE INDEX idx_embeddings_vector ON public.embeddings USING ivfflat (vector) WITH (lists='100');

alter table "public"."chunks" add constraint "chunks_pkey" PRIMARY KEY using index "chunks_pkey";

alter table "public"."embeddings" add constraint "embeddings_pkey" PRIMARY KEY using index "embeddings_pkey";

alter table "public"."chunks" add constraint "chunks_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES lectures(id) ON DELETE CASCADE not valid;

alter table "public"."chunks" validate constraint "chunks_lecture_id_fkey";

alter table "public"."chunks" add constraint "chunks_lecture_id_slide_number_chunk_index_key" UNIQUE using index "chunks_lecture_id_slide_number_chunk_index_key";

alter table "public"."chunks" add constraint "fk_chunks_slide" FOREIGN KEY (lecture_id, slide_number) REFERENCES slides(lecture_id, slide_number) ON DELETE CASCADE not valid;

alter table "public"."chunks" validate constraint "fk_chunks_slide";

alter table "public"."embeddings" add constraint "embeddings_lecture_id_fkey" FOREIGN KEY (lecture_id) REFERENCES lectures(id) ON DELETE CASCADE not valid;

alter table "public"."embeddings" validate constraint "embeddings_lecture_id_fkey";

alter table "public"."embeddings" add constraint "fk_embeddings_chunk" FOREIGN KEY (chunk_id) REFERENCES chunks(id) ON DELETE CASCADE not valid;

alter table "public"."embeddings" validate constraint "fk_embeddings_chunk";

grant delete on table "public"."chunks" to "anon";

grant insert on table "public"."chunks" to "anon";

grant references on table "public"."chunks" to "anon";

grant select on table "public"."chunks" to "anon";

grant trigger on table "public"."chunks" to "anon";

grant truncate on table "public"."chunks" to "anon";

grant update on table "public"."chunks" to "anon";

grant delete on table "public"."chunks" to "authenticated";

grant insert on table "public"."chunks" to "authenticated";

grant references on table "public"."chunks" to "authenticated";

grant select on table "public"."chunks" to "authenticated";

grant trigger on table "public"."chunks" to "authenticated";

grant truncate on table "public"."chunks" to "authenticated";

grant update on table "public"."chunks" to "authenticated";

grant delete on table "public"."chunks" to "service_role";

grant insert on table "public"."chunks" to "service_role";

grant references on table "public"."chunks" to "service_role";

grant select on table "public"."chunks" to "service_role";

grant trigger on table "public"."chunks" to "service_role";

grant truncate on table "public"."chunks" to "service_role";

grant update on table "public"."chunks" to "service_role";

grant delete on table "public"."embeddings" to "anon";

grant insert on table "public"."embeddings" to "anon";

grant references on table "public"."embeddings" to "anon";

grant select on table "public"."embeddings" to "anon";

grant trigger on table "public"."embeddings" to "anon";

grant truncate on table "public"."embeddings" to "anon";

grant update on table "public"."embeddings" to "anon";

grant delete on table "public"."embeddings" to "authenticated";

grant insert on table "public"."embeddings" to "authenticated";

grant references on table "public"."embeddings" to "authenticated";

grant select on table "public"."embeddings" to "authenticated";

grant trigger on table "public"."embeddings" to "authenticated";

grant truncate on table "public"."embeddings" to "authenticated";

grant update on table "public"."embeddings" to "authenticated";

grant delete on table "public"."embeddings" to "service_role";

grant insert on table "public"."embeddings" to "service_role";

grant references on table "public"."embeddings" to "service_role";

grant select on table "public"."embeddings" to "service_role";

grant trigger on table "public"."embeddings" to "service_role";

grant truncate on table "public"."embeddings" to "service_role";

grant update on table "public"."embeddings" to "service_role";


