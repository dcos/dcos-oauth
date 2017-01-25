package main

import (
	"github.com/dcos/dcos-oauth/common"
	"net/http"
	"golang.org/x/net/context"
	"log"
	"github.com/coreos/go-oidc/oauth2"
	"time"
	"encoding/json"
	"io/ioutil"
	"encoding/base64"
)

func oauth2Client(ctx context.Context) *oauth2.Client {
	key := ctx.Value("oauth-app-key").(string)
	secret := ctx.Value("oauth-app-secret").(string)
	tokenUrl := ctx.Value("oauth-token-url").(string)
	authUrl := ctx.Value("oauth-auth-url").(string)
	callbackUrl := ctx.Value("oauth-callback-url").(string)

	conf := oauth2.Config{
		Credentials: oauth2.ClientCredentials{ID: key, Secret: secret},
		Scope:       []string{"foo-scope", "bar-scope"},
		TokenURL:    tokenUrl,
		AuthMethod:  oauth2.AuthMethodClientSecretBasic,
		RedirectURL: callbackUrl,
		AuthURL:     authUrl,
	}

	o2cli, _ := oauth2.NewClient(httpClient, conf)
	return o2cli
}

func handleCallback(ctx context.Context, w http.ResponseWriter, r *http.Request) *common.HttpError {
	code := r.URL.Query()["code"]
	o2cli := oauth2Client(ctx)
	token, _ := o2cli.RequestToken(oauth2.GrantTypeAuthCode, code[0])
	client := &http.Client{}

	profileUrl := ctx.Value("oauth-profile-url").(string)+token.AccessToken
	req, err := http.NewRequest("GET", profileUrl, nil)

	resp, err := client.Do(req)
	if err!=nil{
		log.Print("error %w",err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	userJson := string(contents)

	const cookieMaxAge = 388800
	// required for IE 6, 7 and 8
	expiresTime := time.Now().Add(cookieMaxAge * time.Second)

	authCookie := &http.Cookie{
		Name:     "dcos-acs-auth-cookie",
		Value:    token.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Expires:  expiresTime,
		MaxAge:   cookieMaxAge,
	}
	http.SetCookie(w, authCookie)

	user := User{
		Uid:         userJson,
		Description: userJson,
		IsRemote:    false,
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Printf("Marshal: %v", err)
		return common.NewHttpError("JSON marshalling failed", http.StatusInternalServerError)
	}
	infoCookie := &http.Cookie{
		Name:    "dcos-acs-info-cookie",
		Value:   base64.URLEncoding.EncodeToString(userBytes),
		Path:    "/",
		Expires: expiresTime,
		MaxAge:  cookieMaxAge,
	}

	http.SetCookie(w, infoCookie)
	json.NewEncoder(w).Encode(loginResponse{Token: string(userJson)})
	return nil
}