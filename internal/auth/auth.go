package auth

const (
	defaultClientID = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
	defaultTenantID = "organizations"
	defaultScope    = "https://management.core.windows.net//.default offline_access"
)

func GetToken() (string, error) {
	if token, ok := tryCache(); ok {
		return token, nil
	}

	return login()
}

func tryCache() (string, bool) {
	data, err := loadCache()
	if err != nil {
		return "", false
	}

	if !data.IsExpired() {
		return data.AccessToken, true
	}

	if data.RefreshToken == "" {
		return "", false
	}

	tenant := data.TenantID
	if tenant == "" {
		tenant = defaultTenantID
	}

	tr, err := refreshToken(data.RefreshToken, tenant)
	if err != nil {
		return "", false
	}

	if tenant == defaultTenantID {
		if realTenant, err := getTenant(tr.AccessToken); err == nil {
			if tr2, err := refreshToken(tr.RefreshToken, realTenant); err == nil {
				tr = tr2
				tenant = realTenant
			}
		}
	}

	_ = saveCache(tr, tenant)

	return tr.AccessToken, true
}

func login() (string, error) {
	tr, err := deviceLogin()
	if err != nil {
		return "", err
	}

	tenant := defaultTenantID

	if realTenant, err := getTenant(tr.AccessToken); err == nil {
		if tr2, err := refreshToken(tr.RefreshToken, realTenant); err == nil {
			tr = tr2
			tenant = realTenant
		}
	}

	_ = saveCache(tr, tenant)

	return tr.AccessToken, nil
}
