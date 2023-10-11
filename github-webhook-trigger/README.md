# Github webhook trigger

The github webhook trigger is a predefined KAI process that creates a github webhook passive listener. It will stay listening for user requested events, then trigger a KAI workflow on said event.  
The trigger supports the following event types:

- push
- pull
- release
- workflow_dispatch
- workflow_run

## How to setup

The trigger requires adding two configuration options to the process-scoped configuration.
One being the events the webhook will listen to (_webhook_events_), the other the github secret needed to interact with the github repo (_github_secret_).

### Configuration

| Key            | Optional  | Type | Value                                                                                         |
|----------------|-----------|------|-----------------------------------------------------------------------------------------------|
| webhook_events | no        | str  | Possible options (comma separated): push, pull, release, workflow_dispatch, workflow_run      |
| github_secret  | yes       | str  | Github's repository secret.  |

! Github repository needs to be configured also to expose events to `/webhook-github` please check [webhook_guide](https://docs.github.com/webhooks/) for more information.

### Output

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------|
| requestID | str  | A string uuid                                                          |
| eventUrl  | str  | The url from the repo triggering the event                             |
| event     | str  | The name of the event that has occurred                                |

The url can be defined in `https://github.com/<YOUR_REPOSITORY>/settings/hooks`

#### Example

```json
{
 "requestID: 123
 "eventUrl": http:://example/webhook-github
 "event": push
}
```
