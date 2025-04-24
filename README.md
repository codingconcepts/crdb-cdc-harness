# crdb-cdc-harness
A little test harness that shows latencies and message counts for rows emitted by CockroachDB changefeeds.

### Installation

Find the release that matches your architecture on the [releases](https://github.com/codingconcepts/crdb-cdc-harness/releases) page.

Download the tar for your OS and architecture, extract the executable, and move it into your PATH. For example:

```sh
tar -xvf cdch_[VERSION]_[OS]_[ARCH].tar.gz
```

### Usage Example

Help text

```sh
cdch

Usage of cdch:
  -arg value
        argument generator
  -database-url string
        database connection string
  -kafka-message-id-path string
        dot notation path to the id in the kafka message
  -kafka-topic string
        name of the kafka topic
  -kafka-url string
        url to the kafka broker
  -latency-pool-size int
        average latency sample size (default 100)
  -version
        display the application version
  -write-frequency duration
        database write frequency (default 1s)
  -write-statement string
        database write statement to use
```

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

Run cdch

```sh
cdch \
--database-url "postgres://root@localhost:26257?sslmode=disable" \
--kafka-url localhost:9092 \
--kafka-topic purchase \
--kafka-message-id-path after.id \
--write-frequency 10ms \
--latency-pool-size 100 \
--write-statement 'INSERT INTO purchase (customer_id, total) VALUES ($1, $2) RETURNING id' \
--arg uuid \
--arg price
```

Teardown

```sh
(cd example && docker compose down)
```