ALTER TABLE public.refresh_tokens ALTER COLUMN "token" TYPE varchar(256) USING "token"::varchar;
ALTER TABLE public.users ALTER COLUMN "hash" TYPE varchar(256) USING "hash"::varchar;
ALTER TABLE public.users ALTER COLUMN "name" TYPE varchar(200) USING "name"::varchar;
ALTER TABLE public.users ALTER COLUMN "email" TYPE varchar(320) USING "email"::varchar;
