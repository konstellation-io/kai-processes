package webhook

import (
	"errors"
	"fmt"
)

var ErrEventNotSupported = errors.New("event not supported")
var ErrNotAValidEvent = errors.New("not a valid event")
var ErrGettingEventsFromConfig = errors.New("error getting events from config")
var ErrCreatingWebhook = errors.New("error creating webhook")
var ErrServer = errors.New("error listening and serving")

func concatenateErrors(err1, err2 error) error {
	return fmt.Errorf("%w: %w", err1, err2)
}

func NotValidEventError(event string) error {
	return fmt.Errorf("%q is %w", event, ErrNotAValidEvent)
}

func GettingEventsFromConfigError(err error) error {
	return concatenateErrors(ErrGettingEventsFromConfig, err)
}

func CreatingWebhookError(err error) error {
	return concatenateErrors(ErrCreatingWebhook, err)
}

func ServerError(err error) error {
	return concatenateErrors(ErrServer, err)
}
