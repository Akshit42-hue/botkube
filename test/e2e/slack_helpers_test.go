//go:build integration

package e2e

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var slackLinks = regexp.MustCompile(`<(?P<val>https://[^>]*)>`)

func removeSlackLinksIndicators(content string) string {
	tpl := "$val"

	return slackLinks.ReplaceAllStringFunc(content, func(s string) string {
		var result []byte
		result = slackLinks.ExpandString(result, tpl, s, slackLinks.FindSubmatchIndex([]byte(s)))
		return string(result)
	})
}

func TestRemoveSlackLinksIndicators(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{

		{
			name:     "no links",
			content:  `"...and now my watch begins for cluster '%s'! :crossed_swords:"`,
			expected: `"...and now my watch begins for cluster '%s'! :crossed_swords:"`,
		},
		{
			name:     "one link",
			content:  `You can extend BotKube functionality by writing additional filters that can check resource specs, validate some checks and add messages to the Event struct. Learn more at <https://botkube.io/filters>`,
			expected: `You can extend BotKube functionality by writing additional filters that can check resource specs, validate some checks and add messages to the Event struct. Learn more at https://botkube.io/filters`,
		},
		{
			name:     "multiple links",
			content:  `Learn more at <https://botkube.io/filters> and <https://botkube.io/source>`,
			expected: `Learn more at https://botkube.io/filters and https://botkube.io/source`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// when
			got := removeSlackLinksIndicators(tc.content)

			// then
			assert.Equal(t, tc.expected, got)
		})
	}
}
