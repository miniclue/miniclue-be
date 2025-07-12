alter table "public"."slide_images" alter column "type" set data type text using "type"::text;

drop type "public"."slide_image_type";


