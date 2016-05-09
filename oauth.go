package goshopify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
)

// Returns a Shopify oauth authorization url for the given shopname and state.
//
// State is a unique value that can be used to check the authenticity during a
// callback from Shopify.
func (app App) AuthorizeUrl(shopName string, state string) string {
	shopUrl, _ := url.Parse(ShopBaseUrl(shopName))
	shopUrl.Path = "/admin/oauth/authorize"
	query := shopUrl.Query()
	query.Set("client_id", app.ApiKey)
	query.Set("redirect_uri", app.RedirectUrl)
	query.Set("scope", app.Scope)
	query.Set("state", state)
	shopUrl.RawQuery = query.Encode()
	return shopUrl.String()
}

func (app App) GetAccessToken(shopName string, code string) (string, error) {
	type Token struct {
		Token string `json:"access_token"`
	}

	data := struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
	}{
		ClientId:     app.ApiKey,
		ClientSecret: app.ApiSecret,
		Code:         code,
	}

	client := NewClient(app, shopName, "")
	req, err := client.NewRequest("POST", "admin/oauth/access_token", data)

	token := new(Token)
	err = client.Do(req, token)
	return token.Token, err
}

// Verify a message against a message HMAC
func (app App) VerifyMessage(message, messageMAC string) bool {
	mac := hmac.New(sha256.New, []byte(app.ApiSecret))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)

	// shopify HMAC is in hex so it needs to be decoded
	actualMac, _ := hex.DecodeString(messageMAC)

	return hmac.Equal(actualMac, expectedMAC)
}

// Verify callback authorization for the given shop.
func (app App) VerifyAuthorization(shop, code, timestamp, messageMAC string) bool {
	// url values automatically sorts the query string parameters.
	v := url.Values{}
	v.Set("shop", shop)
	v.Set("code", code)
	v.Set("timestamp", timestamp)
	message := v.Encode()

	return app.VerifyMessage(message, messageMAC)
}

// Convenience function for verifying a URL directly from a handler.
func (app App) VerifyAuthorizationURL(u *url.URL) bool {
	q := u.Query()
	return app.VerifyAuthorization(q.Get("shop"), q.Get("code"), q.Get("timestamp"), q.Get("hmac"))
}
