# Kafka trigger

The Kafka trigger is a predefined KAI process that will stand by listening to a Kafka topic, for every message sent to the topic the
trigger will start the workflow.

The trigger will output the topic's name and the event's data.

## How to setup

The trigger uses five configuration variables, two of them are optional:

- The addresses in where the brokers are available (a single string comma separated).
- The group ID desired to be used.
- The topic's name the trigger will be listening.
- Enable TLS connection (optional).
- Skip verification of trigger's SSL certificates (optional).

### Configuration

The configuration should be defined inside the `centralized configuration scope`:

| Key         | Optional  | Type | Value                                       |
|-------------|-----------|------|---------------------------------------------|
| brokers     | no        | str  | Brokers' addresses (comma separated value)  |
| groupid     | no        | str  | The groupID the listener will take          |
| topic       | no        | str  | The topic's name                            |
| tls_enabled | yes        | bool  | Enable TLS connection, defaults to false  |
| skip_tls_verify | yes        | bool  | Skip SSL certificate validation, defaults to false  |

#### Configuration example

``` yaml
centralized_configuration:
  config:
    brokers: 'localhost:29092'
    groupid: 'kafka-group-id'
    topic: 'test'
    tls_enabled: true
    skip_tls_verify: true
```

### Output

The trigger upon receiving a message, will send a key-value protobuf message with the following format:

| Key       | Type    | Value                    |
|-----------|---------|--------------------------|
| topic     | str     | The topic's name         |
| payload   | []byte  | The event's payload    |

#### Output example

```json
{
 "topic": "test-topic",
 "payload": {"json": "test"}
}
```
