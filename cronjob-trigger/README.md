# Cronjob trigger

The cronjob trigger is a predefined KAI process that will trigger an action in a predefined time.  

## How to setup

The trigger requires adding two configuration options:
- A predefined cron time with the structure `cron: * * * * * *` which equals to seconds | minutes | hours | days | months | years
- The message to be sent

### Config 

The configuration should be defined inside the `centralized configuration scope`, for example:

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

