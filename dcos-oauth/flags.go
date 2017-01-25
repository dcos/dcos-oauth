package main

import "github.com/codegangsta/cli"

var (
	// TODO can we proxy this?
	flIssuerURL = cli.StringFlag{
		Name:   "issuer-url",
		Usage:  "JWT Issuer URL",
		Value:  "https://dcos.auth0.com/",
		EnvVar: "OAUTH_ISSUER_URL",
	}

	flClientID = cli.StringFlag{
		Name:   "client-id",
		Usage:  "JWT Client ID",
		Value:  "3yF5TOSzdlI45Q1xspxzeoGBe9fNxm9m",
		EnvVar: "OAUTH_CLIENT_ID",
	}

	flSecretKeyPath = cli.StringFlag{
		Name:   "secret-key-path",
		Usage:  "Secret key file path",
		Value:  "/var/lib/dcos/auth-token-secret",
		EnvVar: "SECRET_KEY_FILE_PATH",
	}

	flSegmentKey = cli.StringFlag{
		Name:  "segment-key",
		Usage: "Segment key",
		Value: "39uhSEOoRHMw6cMR6st9tYXDbAL3JSaP",
	}
	flProtocol = cli.StringFlag{
		Name:  "protocol",
		Usage: "protocol",
		Value: "oid",
	}
	flOauthAppKey = cli.StringFlag{
		Name: "oauth-app-key",
		Usage: "app key",
		Value: "myApp key",
	}
	flOauthAppSecret = cli.StringFlag{
		Name: "oauth-app-secret",
		Usage: "app secret",
		Value: "myApp secret",
	}
	flOauthTokenUrl = cli.StringFlag{
		Name: "oauth-token-url",
		Usage: "oauth-token-url",
		Value: "https://oauth-token-url",
	}
	flOauthAuthUrl = cli.StringFlag{
		Name: "oauth-auth-url",
		Usage: "oauth-auth-url",
		Value: "https://oauth-auth-url",
	}
	flOauthCallbackUrl = cli.StringFlag{
		Name: "oauth-callback-url",
		Usage: "oauth-callback-url",
		Value: "https://oauth-callback-url",
	}
	flOauthProfileUrl = cli.StringFlag{
		Name: "oauth-profile-url",
		Usage: "profile url",
		Value: "https://oauth/profile?access_token=",
	}
)
