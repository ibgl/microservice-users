ALTER TABLE public.refresh_tokens ALTER COLUMN "token" TYPE varchar USING "token"::varchar;
ALTER TABLE public.users ALTER COLUMN "hash" TYPE varchar USING "hash"::varchar;
ALTER TABLE public.users ALTER COLUMN "name" TYPE varchar USING "name"::varchar;
ALTER TABLE public.users ALTER COLUMN "email" TYPE varchar USING "email"::varchar;