//go:build !dev
// +build !dev

package clients

// OidcLogin attempts to login to Conjur using the OIDC flow. Username and password are ignored - they are
// only used for testing (see the dev build tag - authn_oidc_dev.go)
func OidcLogin(conjurClient ConjurClient, username string, password string) (ConjurClient, error) {
	return oidcLogin(conjurClient, openBrowser)
}
