# GRPC trigger

The GRPC webhook trigger is a predefined KAI process that creates a GRPC server instance.  
It will stay listening for client request through port 8080.  

## How to setup

There is no required configuration in your version file to use this process.

However, processes communicating with this trigger must be prepared to do so. Processes listening to
the grpc trigger need to unmmarshal protobuf based JSON messages, and processes sending messages to
the trigger need to marshal protobuf based JSON messages.

Both output and input messages towards this process require the usage of specific keys in the JSON.

To start communicating with the trigger once it is exposed it is recommended to do a reflection on the GRPC server. You can also download the proto file found within this repository and use it.

### Output (JSON)

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------|
| param1    | str  | A param sent by the user                                               |
| param2    | str  | A param sent by the user                                               |
| param3    | str  | A param sent by the user                                               |

#### Output Example

```json
{
 "param1": "example param",
 "param2": "a repo's url",
 "param3": "action to make"
}
```

### Input (JSON)

| Key         | Type | Value                                                                  |
|-------------|------|------------------------------------------------------------------------|
| status_code | str  | Status code for the workflow execution                                 |
| message     | str  | A message sent by the user                                             |

#### Input Example

```json
{
 "status_code": "200",
 "message": "All good!",
}
```
