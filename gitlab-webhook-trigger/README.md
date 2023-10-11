# Gitlab webhook trigger

The gitlab webhook trigger is a predefined KAI process that creates a gitlab webhook passive listener. It will stay listening for user requested events, then trigger a KAI workflow on said event.  
The trigger supports the following event types:

- push
- merge_request
- comment
- tag

## How to setup

The trigger requires adding two configuration options to the process-scoped configuration.
One being the events the webhook will listen to (_webhook_events_), the other the gitlab secret needed to interact with the gitlab repo (_gitlab_secret_).

! Gitlab repository needs to be configured also to expose events to the webhook please check [webhook_guide](https://docs.gitlab.com/ee/user/project/integrations/webhooks.html) for more information.

### Configuration

| Key            | Optional  | Type | Value                                                                                         |
|----------------|-----------|------|-----------------------------------------------------------------------------------------------|
| webhook_events | no        | str  | Possible options (comma separated): push, merge_request, comment, tag   |
| gitlab_secret  | yes       | str  | Gitlab's webhook secret |

### Output (JSON)

| Key       | Type | Value                                                                  |
|-----------|------|------------------------------------------------------------------------|
| requestID | str  | A string uuid                                                          |
| eventUrl  | str  | The url from the repo triggering the event                             |
| event     | str  | The name of the event that has occurred                                |

#### Example

```json
{
 "requestID": "123",
 "eventUrl": "http:://example/webhook-github",
 "event": "Push Hook"
}
```
