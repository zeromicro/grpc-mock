package internal

type Config struct {
	MockEnableKey     string `json:",default=mock"`
	MockEnableValue   string `json:",default=yes"`
	MockDisableValue  string `json:",default=no"`
	MockCaseKey       string `json:",default=case_name"`
	MockCustomCaseKey string `json:",default=custom_case"`
}
