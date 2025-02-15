package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMattermost_FindAndTrimBotMention(t *testing.T) {
	/// given
	botName := "BotKube"
	testCases := []struct {
		Name               string
		Input              string
		ExpectedTrimmedMsg string
		ExpectedFound      bool
	}{
		{
			Name:               "Mention",
			Input:              "@BotKube get pods",
			ExpectedFound:      true,
			ExpectedTrimmedMsg: " get pods",
		},
		{
			Name:               "Lowercase",
			Input:              "@botkube get pods",
			ExpectedFound:      true,
			ExpectedTrimmedMsg: " get pods",
		},
		{
			Name:               "Yet another different casing",
			Input:              "@BOTKUBE get pods",
			ExpectedFound:      true,
			ExpectedTrimmedMsg: " get pods",
		},
		{
			Name:          "Not at the beginning",
			Input:         "Not at the beginning @BotKube get pods",
			ExpectedFound: false,
		},
		{
			Name:          "Different mention",
			Input:         "@bootkube get pods",
			ExpectedFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			botMentionRegex, err := mattermostBotMentionRegex(botName)
			require.NoError(t, err)
			b := &Mattermost{botMentionRegex: botMentionRegex}
			require.NoError(t, err)

			// when
			actualTrimmedMsg, actualFound := b.findAndTrimBotMention(tc.Input)

			// then
			assert.Equal(t, tc.ExpectedFound, actualFound)
			assert.Equal(t, tc.ExpectedTrimmedMsg, actualTrimmedMsg)
		})
	}
}
