# Process trigger

The process trigger is a predefined KAI process that will trigger an action when a new event is received on the given subject

## How to setup

The trigger requires adding two configuration options:
- A string field process which the subject to subscribe
- A boolean field retain-execution-id

### Configuration 

The configuration should be defined inside the `centralized configuration scope`:

| Key            | Optional  | Type | Value                                                                                         |
|----------------|-----------|------|-----------------------------------------------------------------------------------------------|
| process | no        | str  | Subject to subscribe     |
| message | yes        | bool  | Default value is true      |

#### Example

```
centralized_configuration:
  process:
    bucket: process
    config:
      process: process
      retain-execution-id: true
```

### Output

It triggers an event of sending a message through the module messaging in the sdk.

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------|
| requestID | str  | A string                                     |
| message  | str  | The defined message if any    |

#### Example

```
{
	"requestID: 123
	"message": "test message"
}
```

