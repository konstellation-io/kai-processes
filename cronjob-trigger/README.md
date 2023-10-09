# Cronjob trigger

The cronjob trigger is a predefined KAI process that will trigger an action in a predefined time.  

## How to setup

The trigger requires adding two configuration options:
- A predefined cron time with the structure `cron: * * * * * *` which equals to seconds | minutes | hours | days | months | years
- The message to be sent

### Input 

It doesn't accept inputs.

## Output

It triggers an event of sending a message through the module messaging in the sdk in the predefined interval of time.

