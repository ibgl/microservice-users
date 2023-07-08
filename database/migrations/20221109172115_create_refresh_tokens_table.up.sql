CREATE TABLE public.refresh_tokens (
	uuid uuid NOT NULL,
	"token" varchar NOT NULL,
	user_uuid uuid NOT NULL,
	ip varchar NOT NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT refresh_tokens_pk PRIMARY KEY (uuid),
	CONSTRAINT refresh_tokens_fk FOREIGN KEY (user_uuid) REFERENCES public.users(uuid)
);
