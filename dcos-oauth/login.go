package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/go-oidc/jose"
	"github.com/coreos/go-oidc/oidc"
	"github.com/samuel/go-zookeeper/zk"

	"github.com/dcos/dcos-oauth/common"
)

type loginRequest struct {
	Uid      string `json:"uid,omitempty"`

	Password string `json:"password,omitempty"`

	Token    string `json:"token,omitempty"`
}

type loginResponse struct {
	Token string `json:"token,omitempty"`
}


func handleLogin(ctx context.Context, w http.ResponseWriter, r *http.Request) *common.HttpError {
	protocol := ctx.Value("protocol")
	if protocol == "oauth2" {
		o2cli:=oauth2Client(ctx)
		o2AuthUrl := o2cli.AuthCodeURL("d", "post", "")
		log.Printf(o2AuthUrl)
		http.Redirect(w,r,o2AuthUrl,302)
	} else {
		var lr loginRequest
		err := json.NewDecoder(r.Body).Decode(&lr)
		if err != nil {
			log.Printf("Decode: %v", err)
			return common.NewHttpError("JSON decode error", http.StatusBadRequest)
		}

		issuerURL, _ := ctx.Value("issuer-url").(string)
		log.Printf("issuerURL g: %v", issuerURL)

		provCfg, err := oidc.FetchProviderConfig(httpClient, issuerURL)
		if err != nil {
			log.Printf("FetchProviderConfig: %v", err)
			return common.NewHttpError("[OIDC] Fetch provider config error", http.StatusInternalServerError)
		}

		clientID, _ := ctx.Value("client-id").(string)

		cliCfg := oidc.ClientConfig{
			HTTPClient:     httpClient,
			ProviderConfig: provCfg,
			Credentials: oidc.ClientCredentials{
				ID: clientID,
			},
		}
		oidcCli, err := oidc.NewClient(cliCfg)
		if err != nil {
			log.Printf("oidc.NewClient: %v", err)
			return common.NewHttpError("[OIDC] Client creation error", http.StatusInternalServerError)
		}

		token, err := jose.ParseJWT(lr.Token)
		if err != nil {
			log.Printf("ParseJWT: %v", err)
			return common.NewHttpError("JWT parsing failed", http.StatusBadRequest)
		}

		err = oidcCli.VerifyJWT(token)
		if err != nil {
			log.Printf("VerifyJWT: %v", err)
			return common.NewHttpError("JWT verification failed", http.StatusUnauthorized)
		}

		claims, err := token.Claims()
		if err != nil {
			log.Printf("Claims: %v", err)
			return common.NewHttpError("invalid claims", http.StatusBadRequest)
		}

		// check for Auth0 email verification
		if verified, ok := claims["email_verified"]; ok {
			if b, ok := verified.(bool); ok && !b {
				log.Printf("email not verified")
				return common.NewHttpError("email not verified", http.StatusBadRequest)
			}
		}

		uid, ok, err := claims.StringClaim("email")
		if !ok || err != nil {
			return common.NewHttpError("invalid email claim", http.StatusBadRequest)
		}

		c := ctx.Value("zk").(*zk.Conn)

		users, _, err := c.Children("/dcos/users")
		if err != nil && err != zk.ErrNoNode {
			return common.NewHttpError("invalid email", http.StatusInternalServerError)
		}

		userPath := fmt.Sprintf("/dcos/users/%s", uid)
		if len(users) == 0 {
			// create first user
			log.Printf("creating first user %v", uid)
			err = common.CreateParents(c, userPath, []byte(uid))
			if err != nil {
				return common.NewHttpError("Zookeeper error", http.StatusInternalServerError)
			}
		}

		exists, _, err := c.Exists(userPath)
		if err != nil || !exists {
			return common.NewHttpError("User unauthorized", http.StatusUnauthorized)
		}

		claims.Add("uid", uid)

		secretKey, _ := ctx.Value("secret-key").([]byte)

		clusterToken, err := jose.NewSignedJWT(claims, jose.NewSignerHMAC("secret", secretKey))
		if err != nil {
			return common.NewHttpError("JWT creation error", http.StatusInternalServerError)
		}
		encodedClusterToken := clusterToken.Encode()

		const cookieMaxAge = 388800
		// required for IE 6, 7 and 8
		expiresTime := time.Now().Add(cookieMaxAge * time.Second)

		authCookie := &http.Cookie{
			Name:     "dcos-acs-auth-cookie",
			Value:    encodedClusterToken,
			Path:     "/",
			HttpOnly: true,
			Expires:  expiresTime,
			MaxAge:   cookieMaxAge,
		}
		http.SetCookie(w, authCookie)

		user := User{
			Uid:         uid,
			Description: uid,
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

		json.NewEncoder(w).Encode(loginResponse{Token: encodedClusterToken})

		return nil
	}
	return nil
}

func handleLogout(ctx context.Context, w http.ResponseWriter, r *http.Request) *common.HttpError {
	// required for IE 6, 7 and 8
	expiresTime := time.Unix(1, 0)

	for _, name := range []string{"dcos-acs-auth-cookie", "dcos-acs-info-cookie"} {
		cookie := &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  expiresTime,
			MaxAge:   -1,
		}

		http.SetCookie(w, cookie)
	}

	return nil
}
