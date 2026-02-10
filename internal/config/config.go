package config

const EnvConfigPrefix = "JIRA_CLAUDE"

type Config struct {
	JiraHost          string `envconfig:"JIRA_HOST" required:"true"`
	JiraUsername      string `envconfig:"JIRA_USERNAME" required:"true"`
	JiraAPIToken      string `envconfig:"JIRA_API_TOKEN" required:"true"`
	BranchPrefix      string `envconfig:"BRANCH_PREFIX" default:"feature/"`
	DefaultBaseBranch string `envconfig:"DEFAULT_BASE_BRANCH" default:"main"`
}
