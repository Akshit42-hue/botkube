package execute

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/kubeshop/botkube/pkg/config"
)

const (
	notifierStartMsgFmt                = "Brace yourselves, incoming notifications from cluster '%s'."
	notifierStopMsgFmt                 = "Sure! I won't send you notifications from cluster '%s' here."
	notifierStatusMsgFmt               = "Notifications from cluster '%s' are %s here."
	notifierNotConfiguredMsgFmt        = "I'm not configured to send notifications here ('%s') from cluster '%s', so you cannot turn them on or off."
	notifierPersistenceNotSupportedFmt = "Platform %q doesn't support persistence for notifications. When BotKube Pod restarts, default notification settings will be applied for this platform."
)

// NotifierHandler handles disabling and enabling notifications for a given communication platform.
type NotifierHandler interface {
	// NotificationsEnabled returns current notification status for a given conversation ID.
	NotificationsEnabled(conversationID string) bool

	// SetNotificationsEnabled sets a new notification status for a given conversation ID.
	SetNotificationsEnabled(conversationID string, enabled bool) error

	BotName() string
}

var (
	// ErrNotificationsNotConfigured describes an error when user wants to toggle on/off the notifications for not configured channel.
	ErrNotificationsNotConfigured = errors.New("notifications not configured for this channel")
)

// NotifierExecutor executes all commands that are related to notifications.
type NotifierExecutor struct {
	log               logrus.FieldLogger
	analyticsReporter AnalyticsReporter
	cfgManager        ConfigPersistenceManager

	// Used for deprecated showControllerConfig function.
	cfg config.Config
}

// NewNotifierExecutor creates a new instance of NotifierExecutor.
func NewNotifierExecutor(log logrus.FieldLogger, cfg config.Config, cfgManager ConfigPersistenceManager, analyticsReporter AnalyticsReporter) *NotifierExecutor {
	return &NotifierExecutor{
		log:               log,
		cfg:               cfg,
		cfgManager:        cfgManager,
		analyticsReporter: analyticsReporter,
	}
}

// Do executes a given Notifier command based on args.
func (e *NotifierExecutor) Do(ctx context.Context, args []string, commGroupName string, platform config.CommPlatformIntegration, conversation Conversation, clusterName string, handler NotifierHandler) (string, error) {
	if len(args) != 2 {
		return "", errInvalidCommand
	}

	var cmdVerb = args[1]
	var isUnknownVerb bool
	defer func() {
		if isUnknownVerb {
			cmdVerb = anonymizedInvalidVerb // prevent passing any personal information
		}
		cmdToReport := fmt.Sprintf("%s %s", args[0], cmdVerb)
		err := e.analyticsReporter.ReportCommand(platform, cmdToReport, conversation.IsButtonClickOrigin)
		if err != nil {
			// TODO: Return error when the DefaultExecutor is refactored as a part of https://github.com/kubeshop/botkube/issues/589
			e.log.Errorf("while reporting notifier command: %s", err.Error())
		}
	}()

	switch NotifierAction(strings.ToLower(cmdVerb)) {
	case Start:
		const enabled = true
		err := handler.SetNotificationsEnabled(conversation.ID, enabled)
		if err != nil {
			if errors.Is(err, ErrNotificationsNotConfigured) {
				return fmt.Sprintf(notifierNotConfiguredMsgFmt, conversation.ID, clusterName), nil
			}

			return "", fmt.Errorf("while setting notifications to %t: %w", enabled, err)
		}

		successMessage := fmt.Sprintf(notifierStartMsgFmt, clusterName)
		err = e.cfgManager.PersistNotificationsEnabled(ctx, commGroupName, platform, conversation.Alias, enabled)
		if err != nil {
			if err == config.ErrUnsupportedPlatform {
				e.log.Warnf(notifierPersistenceNotSupportedFmt, platform)
				return successMessage, nil
			}

			return "", fmt.Errorf("while persisting configuration: %w", err)
		}

		return successMessage, nil
	case Stop:
		const enabled = false
		err := handler.SetNotificationsEnabled(conversation.ID, enabled)
		if err != nil {
			if errors.Is(err, ErrNotificationsNotConfigured) {
				return fmt.Sprintf(notifierNotConfiguredMsgFmt, conversation.ID, clusterName), nil
			}

			return "", fmt.Errorf("while setting notifications to %t: %w", enabled, err)
		}

		successMessage := fmt.Sprintf(notifierStopMsgFmt, clusterName)
		err = e.cfgManager.PersistNotificationsEnabled(ctx, commGroupName, platform, conversation.Alias, enabled)
		if err != nil {
			if err == config.ErrUnsupportedPlatform {
				e.log.Warnf(notifierPersistenceNotSupportedFmt, platform)
				return successMessage, nil
			}

			return "", fmt.Errorf("while persisting configuration: %w", err)
		}

		return successMessage, nil
	case Status:
		enabled := handler.NotificationsEnabled(conversation.ID)

		enabledStr := "enabled"
		if !enabled {
			enabledStr = "disabled"
		}

		return fmt.Sprintf(notifierStatusMsgFmt, clusterName, enabledStr), nil
	case ShowConfig:
		out, err := e.showControllerConfig()
		if err != nil {
			return "", fmt.Errorf("while executing 'showconfig' command: %w", err)
		}

		return fmt.Sprintf("Showing config for cluster %q:\n\n%s", clusterName, out), nil
	default:
		isUnknownVerb = true
	}

	return "", errUnsupportedCommand
}

const redactedSecretStr = "*** REDACTED ***"

// Deprecated: this function doesn't fit in the scope of notifier. It was moved from legacy reasons, but it will be removed in future.
func (e *NotifierExecutor) showControllerConfig() (string, error) {
	cfg := e.cfg

	// hide sensitive info
	// TODO: avoid printing sensitive data without need to resetting them manually (which is an error-prone approach)
	for key, old := range cfg.Communications {
		old.Slack.Token = redactedSecretStr
		old.SocketSlack.AppToken = redactedSecretStr
		old.SocketSlack.BotToken = redactedSecretStr
		old.Elasticsearch.Password = redactedSecretStr
		old.Discord.Token = redactedSecretStr
		old.Mattermost.Token = redactedSecretStr
		old.Teams.AppPassword = redactedSecretStr

		// maps are not addressable: https://stackoverflow.com/questions/42605337/cannot-assign-to-struct-field-in-a-map
		cfg.Communications[key] = old
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
