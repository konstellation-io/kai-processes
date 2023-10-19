# REST trigger

The REST webhook trigger is a predefined KAI process that creates a REST server instance.  
It will stay listening for client requests through port 8080.  

## How to setup

There is no required configuration in your version file to use this process.

However, processes communicating with this trigger must be prepared to do so. Processes listening to
the REST trigger need to unmmarshal protobuf based JSON messages, and processes sending messages to
the trigger need to marshal protobuf based JSON messages.

Both output and input messages towards this process require the usage of specific keys in the JSON.

To communicate with the trigger, make a POST operation with a JSON inside containing the following values:

### Trigger's Input from client (JSON)

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------|
| param1    | str  | A param sent by the user                                               |
| param2    | str  | A param sent by the user                                               |
| param3    | str  | A param sent by the user                                               |

#### Trigger's Input from client example

```json
{
 "param1": "example param",
 "param2": "a repo's url",
 "param3": "action to make"
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

### Trigger's Output (JSON)

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------|
| param1    | str  | A param sent by the user                                               |
| param2    | str  | A param sent by the user                                               |
| param3    | str  | A param sent by the user                                               |

#### Trigger's Output Example

```json
{
 "param1": "example param",
 "param2": "a repo's url",
 "param3": "action to make"
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
