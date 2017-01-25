package main

import (
	"github.com/dcos/dcos-oauth/common"
)

var routes = map[string]map[string]common.Handler{
	"POST": {
		"/acs/api/v1/auth/login": handleLogin,
	},
	"PUT": {
		"/acs/api/v1/users/{uid:.*}": putUsers,
	},
	"GET": {
		"/dcos-metadata/ui-config.json": handleUIConfig,
		"/acs/api/v1/auth/logout":       handleLogout,
		"/acs/api/v1/users":             getUsers,
		"/acs/api/v1/users/{uid:.*}":    getUser,
		"/acs/api/v1/auth/login": 	 handleLogin,
		"/oauth2/callback":     handleCallback,
		"/acs/api/v1/groups":            getGroups,
	},
	"DELETE": {
		"/acs/api/v1/users/{uid:.*}": deleteUsers,
	},
}
