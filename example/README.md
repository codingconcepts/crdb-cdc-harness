### Setup

Start CockroachDB and Kafka

```sh
(cd example && docker compose up -d)
```

Initialize CockroachDB and Kafka

```sh
docker exec -it cockroachdb-0 cockroach init --insecure

rpk topic create purchase
```

Connect to CockroachDB

```sh
cockroach sql --insecure
```

Create table and configure changefeed

```sql
CREATE TABLE IF NOT EXISTS purchase (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  customer_id UUID NOT NULL,
  total DECIMAL NOT NULL,
  ts TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE purchase SPLIT AT
  SELECT rpad(to_hex(prefix::INT), 32, '0')::UUID
  FROM generate_series(0, 16) AS prefix;

SET CLUSTER SETTING kv.rangefeed.enabled = true;

CREATE CHANGEFEED FOR TABLE purchase
  INTO "kafka://kafka:29092?topic_name=purchase"
  WITH kafka_sink_config = '{
    "Flush": {
      "MaxMessages": 1000,
      "Frequency": "100ms"
    },
    "RequiredAcks": "ALL"
  }',
  on_error = 'pause',
  initial_scan = 'no';
```

### Teardown

```sh
(cd example && docker compose down)
```