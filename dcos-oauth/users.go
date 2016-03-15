package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"

	"github.com/dcos/dcos-oauth/common"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

type Users struct {
	Array []*User `json:"array"`
}

type User struct {
	Uid         string `json:"uid,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	IsRemote    bool   `json:"is_remote,omitempty"`

	// a quick hack to allow email notifications

	CreatorUid string `json:"creator_uid,omitempty"`
	ClusterURL string `json:"cluster_url,omitempty"`
}

func getUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) *common.HttpError {
	c := ctx.Value("zk").(common.IZk)
	users, _, err := c.Children("/dcos/users")
	if err != nil && err != zk.ErrNoNode {
		return common.NewHttpError("invalid email", http.StatusInternalServerError)
	}

	// users will be an empty list on ErrNoNode
	var usersJson Users
	for _, user := range users {
		userJson := &User{
			Uid:         user,
			Description: user,
			URL:         "",
			IsRemote:    false,
		}
		usersJson.Array = append(usersJson.Array, userJson)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usersJson)
	log.Debugf("Users listed: %+v\n", users)
	return nil
}

func getUser(ctx context.Context, w http.ResponseWriter, r *http.Request) *common.HttpError {
	// uid is already unescaped here
	uid := mux.Vars(r)["uid"]
	if !common.ValidateEmail(uid) {
		return common.NewHttpError("invalid email", http.StatusInternalServerError)
	}

	c := ctx.Value("zk").(common.IZk)

	path := fmt.Sprintf("/dcos/users/%s", uid)
	exists, _, err := c.Exists(path)
	if err != nil {
		return common.NewHttpError("Zookeeper error", http.StatusInternalServerError)
	}
	if !exists {
		log.Printf("getUser: %v doesn't exist", path)
		return common.NewHttpError("User Not Found", http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	userJson := &User{
		Uid:         uid,
		Description: uid,
		IsRemote:    false,
	}
	json.NewEncoder(w).Encode(userJson)

	log.Debugf("User listed: %+v\n", uid)

	return nil
}

func putUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) *common.HttpError {
	uid := mux.Vars(r)["uid"]
	if !common.ValidateEmail(uid) {
		return common.NewHttpError("invalid email", http.StatusInternalServerError)
	}

	c := ctx.Value("zk").(common.IZk)

	path := fmt.Sprintf("/dcos/users/%s", uid)
	exists, _, err := c.Exists(path)
	if err != nil {
		return common.NewHttpError("Zookeeper error", http.StatusInternalServerError)
	}
	if exists {
		return common.NewHttpError("Already Exists", http.StatusConflict)
	}

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Debugf("putUsers: Decode: %v", err)
		return common.NewHttpError("invalid user json", http.StatusBadRequest)
	}
	log.Printf("user: %+v", user)

	err = common.CreateParents(c, path, []byte(uid))
	if err != nil {
		return common.NewHttpError("Zookeeper error", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusCreated)

	log.Debugf("User created: %+v\n", uid)

	segmentKey := ctx.Value("segment-key").(string)
	go newUserEmail(segmentKey, uid, &user)

	return nil
}

type identifyTraits struct {
	Email string `json:"email"`
}

type identifyRequest struct {
	UserId string `json:"userId"`

	Traits identifyTraits `json:"traits"`
}

type trackProperties struct {
	ParentEmail string `json:"parent_email,omitempty"`
	ClusterURL  string `json:"cluster_url,omitempty"`
	ClusterID   string `json:"clusterId,omitempty"`
}

type trackRequest struct {
	UserId string `json:"userId"`

	Event string `json:"event"`

	Properties trackProperties `json:"properties"`
}

func segmentRequest(segmentKey string, urlStr string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Content-Length", string(len(b)))
	req.SetBasicAuth(segmentKey, "")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// newUserEmail sends an invitiation email with the best known link to the
// cluster using Segment.
func newUserEmail(segmentKey string, uid string, user *User) {
	idReq := identifyRequest{
		UserId: uid,
		Traits: identifyTraits{
			Email: uid,
		},
	}

	err := segmentRequest(segmentKey, "https://api.segment.io/v1/identify", idReq)
	if err != nil {
		log.Infof("newUserEmail: %v", err)
		return
	}

	trackReq := trackRequest{
		UserId: uid,
		Event:  "added_to_cluster",
		Properties: trackProperties{
			ParentEmail: user.CreatorUid,
			ClusterURL:  user.ClusterURL,
			ClusterID:   clusterId(),
		},
	}
	err = segmentRequest(segmentKey, "https://api.segment.io/v1/track", trackReq)
	if err != nil {
		log.Infof("newUserEmail: %v", err)
		return
	}
}

func deleteUsers(ctx context.Context, w http.ResponseWriter, r *http.Request) *common.HttpError {
	uid := mux.Vars(r)["uid"]
	if !common.ValidateEmail(uid) {
		return common.NewHttpError("invalid email", http.StatusInternalServerError)
	}

	c := ctx.Value("zk").(common.IZk)
	path := fmt.Sprintf("/dcos/users/%s", uid)
	exists, _, err := c.Exists(path)
	if err != nil {
		return common.NewHttpError("Zookeeper error", http.StatusInternalServerError)
	}
	if !exists {
		return common.NewHttpError("User not found", http.StatusNotFound)
	}

	err = c.Delete(path, 0)
	if err != nil {
		return common.NewHttpError("Zookeeper error", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
	log.Debugf("User deleted: %+v\n", uid)
	return nil
}
