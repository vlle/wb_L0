CREATE TABLE IF NOT EXISTS delivery (
  id serial PRIMARY KEY,
  name text NOT NULL,
  phone text NOT NULL,
  zip text NOT NULL,
  city text NOT NULL,
  address text NOT NULL,
  region text NOT NULL,
  email text NOT NULL -- with possible domain
);


CREATE TABLE IF NOT EXISTS payment (
  transaction text PRIMARY KEY,
  request_id text,
  currency text NOT NULL,
  provider text NOT NULL,
  amount integer NOT NULL,
  payment_dt bigint NOT NULL,
  bank text NOT NULL,
  delivery_cost integer NOT NULL,
  goods_total integer NOT NULL,
  custom_fee integer NOT NULL
);

CREATE TABLE IF NOT EXISTS item (
  id serial PRIMARY KEY,
  order_id text,
  chrt_id integer NOT NULL,
  track_number text NOT NULL,
  price integer NOT NULL,
  rid text NOT NULL,
  name text NOT NULL,
  sale integer NOT NULL,
  size text NOT NULL,
  total_price integer NOT NULL,
  nm_id integer NOT NULL,
  brand text NOT NULL,
  status integer NOT NULL
);


CREATE TABLE IF NOT EXISTS orders (
  order_uid   text PRIMARY KEY,
  track_number text NOT NULL,
  entry text NOT NULL,

  delivery_id integer NOT NULL,
  order_transaction text NOT NULL,
  FOREIGN KEY (delivery_id) REFERENCES delivery (id),
  FOREIGN KEY (order_transaction) REFERENCES payment (transaction),

  locale text NOT NULL,
  internal_signature text NOT NULL,
  customer_id text NOT NULL,
  delivery_service text NOT NULL,
  shardkey text NOT NULL,
  sm_id integer NOT NULL,
  date_created timestamp with time zone NOT NULL,
  oof_shard text NOT NULL
);
