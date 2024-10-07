-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.withdrawals (
	id int8 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1 NO CYCLE) NOT NULL,
	user_id int8 NOT NULL,
	value numeric(11, 2) NOT NULL,
	created_at timestamp DEFAULT now() NOT NULL,
	order_id int8 NOT NULL,
	CONSTRAINT withdrawals_pk PRIMARY KEY (id)
);
CREATE INDEX withdrawals_user_id_idx ON public.withdrawals USING btree (user_id);


-- public.withdrawn внешние включи

ALTER TABLE public.withdrawals ADD CONSTRAINT withdrawn_orders_fk FOREIGN KEY (order_id) REFERENCES public.orders(id);
ALTER TABLE public.withdrawals ADD CONSTRAINT withdrawn_users_fk FOREIGN KEY (user_id) REFERENCES public.users(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd
