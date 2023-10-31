# Process trigger

The process trigger is a predefined KAI process that will trigger an event when a new event is received on the given subject forwarding the message received

## How to setup

The trigger requires adding the following configuration options:
- A string field product
- A string field version
- A string field workflow
- A string field process
- A boolean field retain-execution-id

The first four fields will be used to create the subscription subject

### Configuration 

The configuration should be defined inside the `centralized configuration scope`:

| Key            | Optional  | Type | Value                                                                                         |
|----------------|-----------|------|-----------------------------------------------------------------------------------------------|
| product | no        | str  | The product in which is the process you want to subscribe to     |
| version | no        | str  | The version in which is the process you want to subscribe to     |
| process | no        | str  | The process in which is the process you want to subscribe to     |
| workflow | no        | str  | The workflow in which is the process you want to subscribe to     |
| retain-execution-id | yes        | bool  | Keeps the original request's id - default value is true      |

#### Subtopic

In order to subscribe to a subtopic, process field needs to be in the form `process.subtopic`

#### Example

```
centralized_configuration:
  process:
    bucket: process
    config:
      product: product
      version: version
      workflow: workflow
      process: process
      retain-execution-id: true
```

### Output

It triggers an event of sending a message through the module messaging in the sdk.

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------|
| requestID | str  | A string                                     |
| message  | str  | The message received    |

#### Example

```
{
	"requestID: 123
	"message": "test message"
}
```

