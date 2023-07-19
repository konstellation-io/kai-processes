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

func NotValidEventError(event string) error {
	return fmt.Errorf("%q is %w", event, ErrNotAValidEvent)
}

func GettingEventsFromConfigError(err error) error {
	return fmt.Errorf("%w: %w", ErrGettingEventsFromConfig, err)
}

func CreatingWebhookError(err error) error {
	return fmt.Errorf("%w: %w", ErrCreatingWebhook, err)
}

func ServerError(err error) error {
	return fmt.Errorf("%w: %w", ErrServer, err)
}