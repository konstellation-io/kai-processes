# Github webhook trigger

The github webhook trigger is a predefined KAI process that creates a webhook listener to a github repository. It will stay listening to user requested events, then trigger a KAI workflow on said event.  
The trigger supports the following event types:

- push
- pull
- release
- workflow_dispatch
- workflow_run

## How to setup

The trigger requires adding two configuration options to the process-scoped configuration.
One being the events the webhook will listen to (_webhook_events_), the other the github secret needed to interact with the github repo (_github_secret_).

### Input 

| Key            | Optional  | Type | Value                                                                                         |
|----------------|-----------|------|-----------------------------------------------------------------------------------------------| 
| webhook_events | no        | str  | Possible options (comma separated): push, pull, release, workflow_dispatch, workflow_run      |
| github_secret  | yes       | str  | Github's repository secret                                                                    |

! Github repository needs to be configured also to expose events to "/webhook-github" please check [webhook_guide](https://docs.github.com/webhooks/) for more information.


## Uploading a Dockerfile to the registry in the local environment

Follow this two-step process:

- Open a Port-forward to the local registry in K9S
- Execute the following command `minikube image build -t <image_name:tag> . -p kai-local`

This will upload the image to a local registry, and will be available to the KAI services.

## Output

```
{
	"eventUrl": url predefined in github
	"event": push, pull, release, workflow_dispatch or workflow_run
}
```
The url can be defined in `https://github.com/<YOUR_REPOSITORY>/settings/hooks`


## Unmarshalling the response

### Json string

Once we get the output we need to convert back from protobuf to JSON, one example in go would be:

```
import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func unmarshalProtobufToJSON(m *structpb.Value) (string, error) {
	if m == nil {
		return "", fmt.Errorf("input protobuf Value is nil")
	}

	// Convert the structpb.Value to a map[string]interface{}
	var data map[string]interface{}
	err := jsonpb.Unmarshal(m, &data)
	if err != nil {
		return "", err
	}

	// Marshal the map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
```

### Go's struct

This method converts the structpb object to a go struct we send as value any

```
func MapStructToStructpb(value any) (*structpb.Struct, error) {
    marshalledUser, err := json.Marshal(val)
    if err != nil {
        return nil, err
    }
    structValue := &structpb.Struct{}
    err = structValue.UnmarshalJSON(marshalledUser)
    if err != nil {
        return nil, err
    }

    return structValue, nil
}
```