# kai-processes
KAI predefined processes

## webhook trigger

This trigger supports the following event types:
- push
- pull
- release
- workflow_dispatch
- workflow_run


## How to setup

The trigger requires two key-values defined in the process' configuration section, inside the krt.yml:

- webhook_events = "event1,event2, event3"
- github_secret = "your_secret"

! Github repository needs to be configured also to expose events to "/webhooks" please check [webhook_guide](https://docs.github.com/webhooks/) for more information.