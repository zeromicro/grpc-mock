package config

type MatchConfig struct {
	MockEnableKey     string `json:",default=mock"`
	MockEnableValue   string `json:",default=yes"`
	MockCaseKey       string `json:",default=case_name"`
	MockCustomCaseKey string `json:",default=custom_case"`
}
