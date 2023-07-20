# kai-processes

KAI predefined processes

## Github webhook trigger

The github webhook trigger is a predefined KIO process that creates a webhook listener to a github repository. It will stay listening to user requested events, then trigger a KAI workflow on said event.  
The trigger supports the following event types:

- push
- pull
- release
- workflow_dispatch
- workflow_run

## How to setup

The trigger requires two key-values defined in the process' configuration section, inside the krt.yml of the KAI product.  
One being the events the webhook will listen to (_webhook_events_), the other the github secret needed to interact with the github repo (_github_secret_).

In example:

- webhook_events = "push, pull, workflow_dispatch"
- github_secret = "your_secret"

! Github repository needs to be configured also to expose events to "/webhooks" please check [webhook_guide](https://docs.github.com/webhooks/) for more information.
