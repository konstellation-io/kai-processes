# Cronjob trigger

The cronjob trigger is a predefined KAI process that will trigger an action in a predefined time.  

## How to setup

The trigger requires adding two configuration options:
- A predefined cron time with the structure `cron: * * * * * *` which equals to seconds | minutes | hours | days | months | years
- The message to be sent

## Input 

The configuration should be defined inside the `centralized configuration scope`:

| Key            | Optional  | Type | Value                                                                                         |
|----------------|-----------|------|-----------------------------------------------------------------------------------------------|
| cron | no        | str  | Cron expression     |
| message | yes        | str  | Message to be sent in the generated events      |

### Example

```
centralized_configuration:
  process:
    bucket: process
    config:
      cron: '30 * * * * *'
      message: 'Hello world'
```

## Output

It triggers an event of sending a message through the module messaging in the sdk in the predefined interval of time.

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------|
| requestID | str  | A string                                     |
| message  | str  | The defined message if any    |
| time     | str  | The timestamp of the generated event |

### Example

```
{
	"requestID: 123
	"message": "test message"
	"time": "Mon Jan 2 15:04:05 MST 2006"
}
```

