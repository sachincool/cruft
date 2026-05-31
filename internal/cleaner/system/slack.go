package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "slack",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "Slack cache (~/Library/Application Support/Slack/Cache). Re-downloads images and metadata as you scroll. Session cookies are NOT touched.",
		Paths: []string{
			"~/Library/Application Support/Slack/Cache",
			"~/Library/Application Support/Slack/Service Worker/CacheStorage",
			"~/Library/Application Support/Slack/Code Cache",
			"~/Library/Application Support/Slack/GPUCache",
		},
		DetectAnyPath: true,
		BusyProcs:     []string{"Slack"},
		Reason:        "Slack desktop cache",
	})
}
