package main

import (
	"encoding/json"
	"net/http"

	"github.com/dcos/dcos-oauth/common"

	"golang.org/x/net/context"
)

type Groups struct {
	Array []*Group `json:"array"`
}

type Group struct {
}

// groups endpoint is used by systemd health check

func getGroups(ctx context.Context, w http.ResponseWriter, r *http.Request) *common.HttpError {
	var groupsJson Groups
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groupsJson)
	return nil
}
