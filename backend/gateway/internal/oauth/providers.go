package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ProviderUser struct {
	ProviderUserID string
	Email          string
	FirstName      string
	LastName       string
	DisplayName    string
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func AuthURL(provider string, cfg ProviderConfig, state string) (string, error) {
	switch strings.ToLower(provider) {
	case "google":
		q := url.Values{}
		q.Set("client_id", cfg.ClientID)
		q.Set("redirect_uri", cfg.RedirectURI)
		q.Set("response_type", "code")
		q.Set("scope", "openid email profile")
		q.Set("state", state)
		q.Set("access_type", "offline")
		q.Set("prompt", "consent")
		return "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode(), nil
	case "github":
		q := url.Values{}
		q.Set("client_id", cfg.ClientID)
		q.Set("redirect_uri", cfg.RedirectURI)
		q.Set("scope", "read:user user:email")
		q.Set("state", state)
		return "https://github.com/login/oauth/authorize?" + q.Encode(), nil
	default:
		return "", fmt.Errorf("unsupported provider")
	}
}

func ExchangeAndFetchUser(httpClient *http.Client, provider string, cfg ProviderConfig, code string) (ProviderUser, error) {
	switch strings.ToLower(provider) {
	case "google":
		return exchangeGoogle(httpClient, cfg, code)
	case "github":
		return exchangeGitHub(httpClient, cfg, code)
	default:
		return ProviderUser{}, fmt.Errorf("unsupported provider")
	}
}

func exchangeGoogle(httpClient *http.Client, cfg ProviderConfig, code string) (ProviderUser, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	form.Set("redirect_uri", cfg.RedirectURI)
	form.Set("grant_type", "authorization_code")

	resp, err := httpClient.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return ProviderUser{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return ProviderUser{}, fmt.Errorf("google token exchange failed")
	}
	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return ProviderUser{}, err
	}
	if token.AccessToken == "" {
		return ProviderUser{}, fmt.Errorf("google access_token missing")
	}

	req, _ := http.NewRequest(http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	uResp, err := httpClient.Do(req)
	if err != nil {
		return ProviderUser{}, err
	}
	defer uResp.Body.Close()
	if uResp.StatusCode >= 300 {
		return ProviderUser{}, fmt.Errorf("google userinfo fetch failed")
	}
	var user struct {
		Sub        string `json:"sub"`
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Name       string `json:"name"`
	}
	if err := json.NewDecoder(uResp.Body).Decode(&user); err != nil {
		return ProviderUser{}, err
	}
	return ProviderUser{
		ProviderUserID: user.Sub,
		Email:          user.Email,
		FirstName:      user.GivenName,
		LastName:       user.FamilyName,
		DisplayName:    user.Name,
	}, nil
}

func exchangeGitHub(httpClient *http.Client, cfg ProviderConfig, code string) (ProviderUser, error) {
	form := url.Values{}
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", cfg.RedirectURI)

	req, _ := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return ProviderUser{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return ProviderUser{}, fmt.Errorf("github token exchange failed")
	}
	var token struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return ProviderUser{}, err
	}
	if token.AccessToken == "" {
		return ProviderUser{}, fmt.Errorf("github access_token missing")
	}

	uReq, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	uReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	uReq.Header.Set("Accept", "application/vnd.github+json")
	uResp, err := httpClient.Do(uReq)
	if err != nil {
		return ProviderUser{}, err
	}
	defer uResp.Body.Close()
	if uResp.StatusCode >= 300 {
		return ProviderUser{}, fmt.Errorf("github user fetch failed")
	}
	var user struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(uResp.Body).Decode(&user); err != nil {
		return ProviderUser{}, err
	}
	email := strings.TrimSpace(user.Email)
	if email == "" {
		eReq, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user/emails", nil)
		eReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
		eReq.Header.Set("Accept", "application/vnd.github+json")
		eResp, err := httpClient.Do(eReq)
		if err != nil {
			return ProviderUser{}, err
		}
		defer eResp.Body.Close()
		if eResp.StatusCode >= 300 {
			body, _ := io.ReadAll(eResp.Body)
			return ProviderUser{}, fmt.Errorf("github emails fetch failed: %s", string(body))
		}
		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := json.NewDecoder(eResp.Body).Decode(&emails); err != nil {
			return ProviderUser{}, err
		}
		for _, e := range emails {
			if e.Primary && e.Verified {
				email = e.Email
				break
			}
		}
		if email == "" && len(emails) > 0 {
			email = emails[0].Email
		}
	}

	firstName := user.Name
	lastName := ""
	if sp := strings.SplitN(strings.TrimSpace(user.Name), " ", 2); len(sp) > 0 {
		firstName = sp[0]
		if len(sp) > 1 {
			lastName = sp[1]
		}
	}
	if firstName == "" {
		firstName = user.Login
	}

	return ProviderUser{
		ProviderUserID: fmt.Sprintf("%d", user.ID),
		Email:          email,
		FirstName:      firstName,
		LastName:       lastName,
		DisplayName:    strings.TrimSpace(user.Name),
	}, nil
}
