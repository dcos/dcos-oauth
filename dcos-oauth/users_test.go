package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"github.com/dcos/dcos-oauth/common"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
)

type MockZk struct {
	path string
}

func (m *MockZk) Children(path string) ([]string, *zk.Stat, error) {
	return []string{""}, nil, nil
}

func (m *MockZk) Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error) {
	return "", nil
}

func (m *MockZk) Delete(path string, version int32) error {
	return nil
}

func (m *MockZk) Exists(path string) (bool, *zk.Stat, error) {
	return true, nil, nil

}

func (m *MockZk) Get(path string) ([]byte, *zk.Stat, error) {
	return []byte{}, nil, nil
}

func TestGetUsers(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	czk := &MockZk{}
	ctx = context.WithValue(ctx, "zk", czk)
	r, _ := http.NewRequest("GET", "/acs/api/v1/users", nil)
	w := httptest.NewRecorder()
	assert.Nil(getUsers(ctx, w, r), "getUsers with valid parameters")

	respBody, _ := ioutil.ReadAll(w.Body)
	assert.Equal("{\"array\":[{\"uid\":\"\",\"description\":\"\",\"is_remote\":false}]}\n", string(respBody), "getUsers body")
}

func TestValidateEmail(t *testing.T) {

	//Note: Our regex support for the following invalid email examples:
	//".email@domain.com",        // Leading dot in address is not allowed
	//"email.@domain.com",        // Trailing dot in address is not allowed
	//"email..email@domain.com",  // Multiple dots
	//"email@-domain.com",        // Leading dash in front of domain is invalid

	badcases := []string{
		"#@%^%#$@#$@#.com",             // Garbage
		"@domain.com",                  // Missing username
		"Test Name <email@domain.com>", // Encoded html within email is invalid
		"email.domain.com",             // Missing @
		"email@domain@domain.com",      // Two @ sign

		"email@domain.com (Test Name)", // Text followed email is not allowed
		"email@domain",                 // Missing top level domain (.com/.net/.org/etc)
		"email@111.222.333.44444 ",     // Invalid IP format
		"nomatching",                   // Missing @ sign and domain
		"email@domain..com",            // Multiple dot in the domain portion is invalid
	}
	for _, example := range badcases {
		if match := common.ValidateEmail(example); match {
			t.Fatalf("For email validation with value: %s, expected: %v, actual: %v", example, false, match)
		}
	}

	goodcases := []string{
		"email@domain.com",              //Valid email
		"firstname.lastname@domain.com", //Email contains dot in the address field
		"email@subdomain.domain.com",    //Email contains dot with subdomain
		"firstname+lastname@domain.com", //Plus sign is considered valid character
		"1234567890@domain.com",         //Digits in address are valid
		"email@domain-one.com",          //Dash in domain name is valid
		"_______@domain.com",            //Underscore in the address field is valid
		"email@domain.co.jp",            //Dot in Top Level Domain name also considered valid (use co.jp as example here)
		"firstname-lastname@domain.com", //Dash in address field is valid
		"email@123.123.123.123",         //Domain is valid IP address
		"email@[123.123.123.123]",       //Square bracket around IP address is considered valid
		"“email”@domain.com",            //Quotes around email is con
	}

	for _, example := range goodcases {
		if match := common.ValidateEmail(example); !match {
			t.Fatalf("For email validation with value: %s, expected: %v, actual: %v", example, true, match)
		}
	}
}
