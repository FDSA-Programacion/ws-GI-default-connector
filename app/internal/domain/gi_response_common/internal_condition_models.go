package giresponsecommon

type InternalCondition struct {
	Message                   string            `json:"message,omitempty"`
	ProviderStatus            string            `json:"providerStatus,omitempty"`
	ProviderStatusDescription string            `json:"providerStatusDescription,omitempty"`
	RsList                    map[string]string `json:"rsList"`
	Status                    string            `json:"status"`
	StatusDescription         string            `json:"statusDescription"`
}

// InternalConditionBook es una estructura específica para Book que permite null en algunos campos
type InternalConditionBook struct {
	Message                   *string           `json:"message"`
	ProviderStatus            *string           `json:"providerStatus"`
	ProviderStatusDescription *string           `json:"providerStatusDescription"`
	RsList                    map[string]string `json:"rsList"`
	Status                    string            `json:"status"`
	StatusDescription         string            `json:"statusDescription"`
}
