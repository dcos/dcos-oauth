package integration

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsers(t *testing.T) {
	assert := assert.New(t)

	assert.NoError(startZk())
	assert.NoError(startOAuthAPI())
	defer cleanup("dcos-oauth")

	exampleEmail := "test@domain.com"
	encoded := url.QueryEscape(exampleEmail)

	getResponse := "{\"array\":[{\"uid\":\"" + exampleEmail + "\",\"description\":\"" + exampleEmail + "\"}]}"

	bodyGetUsers, err := send("GET", "/acs/api/v1/users", 200, nil)
	assert.NoError(err)
	assert.Equal("{\"array\":null}", bodyGetUsers, "Users list should be empty")

	user := struct {
		Uid string `json:"uid"`
	}{exampleEmail}
	_, err = send("PUT", "/acs/api/v1/users/"+encoded, 201, user)
	assert.NoError(err)

	bodyGetUsers, err = send("GET", "/acs/api/v1/users", 200, nil)
	assert.NoError(err)
	assert.Equal(getResponse, bodyGetUsers, "User list should have one address: "+encoded)

	bodyGetUser, err := send("GET", "/acs/api/v1/users/"+encoded, 200, nil)
	assert.NoError(err)
	assert.Equal("{\"uid\":\""+exampleEmail+"\",\"description\":\""+exampleEmail+"\"}", bodyGetUser, "User should return address: test@domain.com")

	_, err = send("DELETE", "/acs/api/v1/users/"+encoded, 204, nil)
	assert.NoError(err)

	bodyGetUsers, err = send("GET", "/acs/api/v1/users", 200, nil)
	assert.NoError(err)
	assert.Equal("{\"array\":null}", bodyGetUsers, "Users list should be empty")

	bodyGetUser, err = send("GET", "/acs/api/v1/users/"+encoded, 404, nil)
	assert.NoError(err)
	assert.Equal("{\"title\":\"Not Found\",\"description\":\"User Not Found\"}", bodyGetUser, "User should be empty")
}
