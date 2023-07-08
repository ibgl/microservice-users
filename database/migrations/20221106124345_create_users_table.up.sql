CREATE TABLE public.users (
	uuid uuid NULL,
	email varchar NOT NULL,
	name varchar NOT NULL,
	hash varchar NOT NULL,
	created_at timestamp NULL,
	updated_at timestamp NULL,
	CONSTRAINT users_pk PRIMARY KEY (uuid)
);

CREATE UNIQUE INDEX users_email_idx ON public.users (email);
