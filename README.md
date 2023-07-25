# kai-processes

KAI predefined processes

## Github webhook trigger

The github webhook trigger is a predefined KAI process that creates a webhook listener to a github repository. It will stay listening to user requested events, then trigger a KAI workflow on said event.  
The trigger supports the following event types:

- push
- pull
- release
- workflow_dispatch
- workflow_run

### How to setup

The trigger requires adding two configuration options to the process-scoped configuration.
One being the events the webhook will listen to (_webhook_events_), the other the github secret needed to interact with the github repo (_github_secret_).

#### Input 

| Key            | Type | Value                                                                       |
|----------------|------|-----------------------------------------------------------------------------| 
| webhook_events | str  | Possible options (comma separated): push, pull, release, workflow_dispatch, workflow_run      |
| github_secret  | str  | Github's repository secret (Optional)                                                    |

! Github repository needs to be configured also to expose events to "/webhooks" please check [webhook_guide](https://docs.github.com/webhooks/) for more information.


### Uploading a Dockerfile to the registry in the local environment

Follow this two-step process:

- Open a Port-forward to the local registry in K9S
- Execute the following command `minikube image build -t <image_name:tag> . -p kai-local`

This will upload the image to a local registry, and will be available to the KAI services.

## Output

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------| 
| RequestID | str  | A string uuid                                                          |
| EventUrl  | str  | An url defined in the github workflow settings                         |
| Event     | str  | Possible options: push, pull, release, workflow_dispatch, workflow_run | 
