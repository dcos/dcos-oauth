package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func startZk() error {
	cmd := exec.Command("docker", "run", "-d", "--net=host", "--name=dcos-zk", "jplock/zookeeper")
	err := cmd.Run()
	if err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func startOAuthAPI() error {
	secretKeyFile, err := ioutil.TempFile("", "dcos-oauth-integration-test")
	if err != nil {
		return err
	}
	defer os.Remove(secretKeyFile.Name())
	cmd := exec.Command("docker", "run", "-d", "-v="+secretKeyFile.Name()+":/var/lib/dcos/auth-token-secret", "--net=host", "--name=dcos-oauth", "dcos-services", "/go/bin/dcos-oauth", "serve")
	err = cmd.Run()
	if err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func startConfigAPI() error {
	cmd := exec.Command("docker", "run", "-d", "--net=host", "--name=dcos-config", "dcos-services", "/go/bin/dcos-config", "serve")
	err := cmd.Run()
	if err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

func cleanup(service string) {
	cmd := exec.Command("docker", "rm", "-f", service, "dcos-zk")
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func encodeData(data interface{}) (*bytes.Buffer, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if err := json.NewEncoder(params).Encode(data); err != nil {
			return nil, err
		}
	}
	return params, nil
}

func send(method, route string, statusExpected int, obj interface{}) (string, error) {
	body, err := encodeData(obj)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(method, "http://127.0.0.1:8101"+route, body)
	if err != nil {
		return "", err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != statusExpected {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("%s", body)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return strings.TrimSpace(string(respBody)), err
}
