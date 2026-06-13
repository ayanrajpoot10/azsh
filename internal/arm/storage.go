package arm

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

func GenerateStorageNames(token string) (accountName, fileShareName string) {
	puid, email := extractClaimsFromToken(token)

	if puid == "" {
		return "", ""
	}

	normalizedEmail := strings.NewReplacer(
		"@", "-",
		".", "-",
		"_", "-",
	).Replace(email)

	accountName = "csg" + puid
	if len(accountName) > 24 {
		accountName = accountName[:24]
	}

	fileShareName = "cs-" + normalizedEmail + "-" + puid

	return accountName, fileShareName
}

func extractClaimsFromToken(token string) (puid, email string) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ""
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", ""
	}

	if p, ok := claims["puid"].(string); ok {
		puid = strings.ToLower(p)
	} else if oid, ok := claims["oid"].(string); ok {
		puid = strings.ToLower(strings.ReplaceAll(oid, "-", ""))
	}

	if e, ok := claims["email"].(string); ok {
		email = e
	} else if un, ok := claims["unique_name"].(string); ok {
		email = un
	}

	return puid, email
}
