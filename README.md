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

### How to setup

The trigger requires adding two configuration options to the process-scoped configuration.
One being the events the webhook will listen to (_webhook_events_), the other the github secret needed to interact with the github repo (_github_secret_).

For example:

- webhook_events = "push, pull, workflow_dispatch"
- github_secret = "your_secret"

! Github repository needs to be configured also to expose events to "/webhooks" please check [webhook_guide](https://docs.github.com/webhooks/) for more information.


### Uploading locally a dockerfile to the registry

For doing this we need two things:

- Doing a Port-forward to the local registry in K9S
- Executing the following command `minikube image build -t <image_name:tag> . -p kai-local`