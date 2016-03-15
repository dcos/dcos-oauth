package common

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

func CreateParents(c IZk, path string, data []byte) error {
	parts := strings.Split(path, "/")

	for i := 1; i <= len(parts); i++ {
		pathParts := strings.Join(parts[:i], "/")
		if pathParts == "" {
			pathParts = "/"
		}

		var b []byte
		if i == len(parts) {
			b = data
		}
		log.Printf("CreateParents: Creating %v", pathParts)

		_, err := c.Create(pathParts, b, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			log.Printf("CreateParents: Create: %v: %v", pathParts, err)
			return err
		}
	}

	return nil
}

func initZk(address, path string) (*zk.Conn, error) {
	czk, _, err := zk.Connect([]string{address}, time.Second)
	if err != nil {
		return nil, err
	}

	err = CreateParents(czk, path, nil)
	if err != nil {
		czk.Close()
		return nil, err
	}

	return czk, nil
}

func ReadLine(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return []byte(strings.TrimSpace(string(b))), nil
}

func ServeCmd(c *cli.Context, ctx context.Context, routes map[string]map[string]Handler) error {
	czk, err := initZk(c.String("zk-addr"), ctx.Value("zk-path").(string))
	if err != nil {
		return err
	}
	defer czk.Close()
	ctx = context.WithValue(ctx, "zk", czk)

	r := NewRouter(ctx, routes)
	return http.ListenAndServe(c.String("addr"), r)
}
