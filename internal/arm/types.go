package arm

type Subscription struct {
	SubscriptionID string `json:"subscriptionId"`
	DisplayName    string `json:"displayName"`
}

type ResourceGroup struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type StorageAccountInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Sku      struct {
		Name string `json:"name"`
	} `json:"sku"`
	Properties struct {
		ProvisioningState string `json:"provisioningState"`
	} `json:"properties"`
}

type Tenant struct {
	ID string `json:"tenantId"`
}

type storageAccountCreatePayload struct {
	Location   string            `json:"location"`
	Sku        map[string]string `json:"sku"`
	Kind       string            `json:"kind"`
	Tags       map[string]string `json:"tags"`
	Properties accountProps      `json:"properties"`
}

type accountProps struct {
	Encryption                encryptionConfig `json:"encryption"`
	SupportsHTTPSOnly         bool              `json:"supportsHttpsTrafficOnly"`
	AllowBlobPublicAccess     bool              `json:"allowBlobPublicAccess"`
	MinimumTLSVersion         string            `json:"minimumTlsVersion"`
}

type encryptionConfig struct {
	Services   encryptionServices `json:"services"`
	KeySource  string             `json:"keySource"`
}

type encryptionServices struct {
	Blob serviceConfig `json:"blob"`
	File serviceConfig `json:"file"`
}

type serviceConfig struct {
	Enabled bool `json:"enabled"`
}
