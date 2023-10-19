# kai-processes

This repo contains the KAI predefined processes.

## How to use

All predefined processes will publish a protobuf compatible structure containing a json inside with key value pairs. Each process will post different json pairs, each explained in their own README.

The process receiving requests from predefined nodes need to be prepared for this format of messages.  
Here are some examples.

### Unmarshalling the response

#### Json string

Once we get the output we need to convert back from protobuf to JSON, one example in go would be:

```go
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

#### Go's struct

This method converts the structpb object to a go struct we send as val any

```go
func MapStructToStructpb(val any) (*structpb.Struct, error) {
    marshalledUser, err := json.Marshal(val)
    if err != nil {
        return nil, err
    }
    structVal := &structpb.Struct{}
    err = structVal.UnmarshalJSON(marshalledUser)
    if err != nil {
        return nil, err
    }

    return structVal, nil
}
```

## Uploading a predefined process to the local environment

If you wish to upload by yourself any predefined process to your local KAI you can do as if it were any other process.  
First clone this repo and from within, by executing the kai command line and uploading the source code, example:

```sh
kli process-registry register trigger github-trigger --dockerfile ./github-webhook-trigger/Dockerfile --product demo --src ./github-webhook-trigger --version v1.0.0
```

This will upload the image to a local registry, and will be available to the KAI services.


## Unpacking proto any to structpb in a process

Given a structpb being defined as:

``` go
m, err := structpb.NewValue(map[string]interface{}{
    "param1": "example"
})
```

which we sent using the sdk with `s.kaiSDK.Messaging.SendOutputWithRequestID(m, reqID)`
or we just got from another third party service, we need first to create the struct to fill with
the given data and then unpack the value to it, which then we can use or transform to another types like JSON:

``` go
var respData struct {
    Param1 string `json:"param1"`
}

responsePb := new(structpb.Value)
err := response.UnmarshalTo(responsePb)
responsePbJSON, err := responsePb.MarshalJSON()
err = json.Unmarshal(responsePbJSON, &respData)
param1 := respData.Param1
```

``` python
    input_proto = Struct()
    response.Unpack(input_proto)

    param1 = input_proto.fields["param1"].string_value
```

