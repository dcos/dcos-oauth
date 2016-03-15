package common

import (
	"net/http"
	"regexp"

	"github.com/samuel/go-zookeeper/zk"
)

type IZk interface {
	Children(path string) ([]string, *zk.Stat, error)
	Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error)
	Delete(path string, version int32) error
	Exists(path string) (bool, *zk.Stat, error)
	Get(path string) ([]byte, *zk.Stat, error)
}

type HttpError struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      int    `json:"-"`
}

func NewHttpError(description string, status int) *HttpError {
	return &HttpError{
		Title:       http.StatusText(status),
		Description: description,
		Status:      status,
	}
}

var (
	emailRe = regexp.MustCompile(`^[a-z0-9“”._%+-]+@(?:[a-z0-9-\[]+\.)+[a-z0-9-\]]{2,}$`)
)

func ValidateEmail(email string) bool {
	return emailRe.MatchString(email)
}
