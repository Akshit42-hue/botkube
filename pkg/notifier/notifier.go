package notifier

import (
	"context"
	"fmt"

	"github.com/kubeshop/botkube/pkg/bot/interactive"
	"github.com/kubeshop/botkube/pkg/config"
	"github.com/kubeshop/botkube/pkg/events"
)

// Notifier sends event notifications and messages on the communication channels.
type Notifier interface {

	// SendEvent notifies about new incoming event from source.
	SendEvent(context.Context, events.Event, []string) error

	// SendMessage is used for notifying about BotKube start/stop listening, possible BotKube upgrades and other events.
	// Some integrations may decide to ignore such messages and have SendMessage method no-op.
	// TODO: Consider option per channel to turn on/off "announcements" (BotKube start/stop/upgrade, notify/config change).
	SendMessage(context.Context, interactive.Message) error

	// IntegrationName returns a name of a given communication platform.
	IntegrationName() config.CommPlatformIntegration

	// Type returns a given integration type. See config.IntegrationType for possible integration types.
	Type() config.IntegrationType
}

// SendPlaintextMessage sends a plaintext message to specified providers.
func SendPlaintextMessage(ctx context.Context, notifiers []Notifier, msg string) error {
	if msg == "" {
		return fmt.Errorf("message cannot be empty")
	}

	// Send message over notifiers
	for _, n := range notifiers {
		err := n.SendMessage(ctx, interactive.Message{
			Base: interactive.Base{
				Body: interactive.Body{
					Plaintext: msg,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("while sending message: %w", err)
		}
	}

	return nil
}
