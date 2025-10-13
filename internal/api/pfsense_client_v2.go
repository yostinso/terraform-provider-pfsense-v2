package pfsense_rest_v2

type TerraformClient struct {
	url         string
	apiUsername string
	apiToken    string
	insecure    bool
	apiClient   *Client
}

func NewTerraformClient(url string, apiUsername string, apiToken string, insecure bool) (*TerraformClient, error) {
	apiClient, error := NewClient(
		url,
	)
	if error != nil {
		return nil, error
	}

	return &TerraformClient{
		url:         url,
		apiUsername: apiUsername,
		apiToken:    apiToken,
		insecure:    insecure,
		apiClient:   apiClient,
	}, nil
}
