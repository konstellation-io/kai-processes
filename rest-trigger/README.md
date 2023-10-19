# REST trigger

The REST webhook trigger is a predefined KAI process that creates a REST server instance.  
It will stay listening for client requests through port 8080.  

## How to setup

There is no required configuration in your version file to use this process.

However, processes communicating with this trigger must be prepared to do so. Processes listening to
the REST trigger need to unmmarshal protobuf based JSON messages, and processes sending messages to
the trigger need to marshal protobuf based JSON messages.

Both output and input messages towards this process require the usage of specific keys in the JSON.

### Trigger's Output (JSON)

| Key       | Type    | Value                                                                  |
|-----------|---------|------------------------------------------------------------------------|
| method    | str     | The REST method used                                                   |
| body      | []byte  | Optional. For POST and PUT methods the body is redirected              |

#### Trigger's Output Example

```json
{
 "method": "POST",
 "body": {"keyA": "valueA", "keyB": "valueB"},
}
```

### Trigger's Input from other processes (JSON)

| Key         | Type | Value                                                                      |
|-------------|------|----------------------------------------------------------------------------|
| status_code | str  | Status code for the workflow execution, used later in http server response |
| message     | str  | A message sent by the user                                                 |

#### Trigger's Input Example

```json
{
 "status_code": "200",
 "message": "All good!",
}
```

### Trigger's Output to client (JSON)

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------|
| message   | str  | A message                                                              |

#### Trigger's Output to client

```json
{
 "message": "All good!",
}
```
