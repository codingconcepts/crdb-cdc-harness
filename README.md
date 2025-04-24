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

The following command runs cdch, pointing at a local CockroachDB cluster and Kafka broker, listening on the purchase topic and writing new purchase events every 10ms. Whenever a message is received from the topic, the row's identifier will be extracted from the "after.id" JSON path.

The insert statement and --arg combinations will generatea new purchase entry, with a randomly generated uuid for the customer_id column and price for the total column.

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
