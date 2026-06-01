package system

import "github.com/sachincool/cruft/internal/cleaner"

func init() {
	// Only the SSO/credential resolution cache (~/.aws/cli/cache). Your
	// config and long-lived credentials (~/.aws/config, ~/.aws/credentials)
	// are never touched. Cached tokens simply re-issue on next call.
	cleaner.Register(&cleaner.PathCleaner{
		NameValue:     "aws-cli",
		CategoryValue: cleaner.CategorySystem,
		Desc:          "AWS CLI credential/SSO resolution cache (~/.aws/cli/cache). Re-issues on next call; config and credentials are untouched.",
		Paths:         []string{"~/.aws/cli/cache"},
		DetectAnyPath: true,
		Reason:        "aws cli token cache",
	})
}
