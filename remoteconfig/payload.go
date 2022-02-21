package remoteconfig

import "time"

const Maintenance = "MAINTENANCE"
const MaintenanceMessage = "MAINTENANCE"

type FbRemoteConfig struct {
	Conditions []struct {
		Name       string `json:"name"`
		Expression string `json:"expression"`
		TagColor   string `json:"tagColor"`
	} `json:"conditions"`
	Version struct {
		VersionNumber string    `json:"versionNumber"`
		UpdateTime    time.Time `json:"updateTime"`
		UpdateUser    struct {
			Email string `json:"email"`
		} `json:"updateUser"`
		UpdateOrigin string `json:"updateOrigin"`
		UpdateType   string `json:"updateType"`
	} `json:"version"`
	ParameterGroups struct {
		VirgoFeatureFlag struct {
			Parameters map[string]VirgoFeatureFlagServices `json:"parameters"`
		} `json:"virgo feature flag"`
	} `json:"parameterGroups"`
}

type VirgoFeatureFlagServices struct {
	DefaultValue struct {
		UseInAppDefault bool `json:"useInAppDefault"`
	} `json:"defaultValue"`
	ConditionalValues map[string]VirgoFeatureFlagEnvironmentStatus `json:"conditionalValues"`
}

type VirgoFeatureFlagEnvironmentStatus struct {
	Value string `json:"value"`
}

type BillPaymentRequest struct {
	ProductType int    `json:"product_type"`
	ProviderID  string `json:"provider_id"`
}