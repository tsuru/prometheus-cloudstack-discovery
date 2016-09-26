package cloudstack

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var dial5Full300ClientNoKeepAlive, _ = makeTimeoutHTTPClient(5*time.Second, 5*time.Minute, -1)

func makeTimeoutHTTPClient(dialTimeout time.Duration, fullTimeout time.Duration, maxIdle int) (*http.Client, *net.Dialer) {
	dialer := &net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: 30 * time.Second,
	}
	client := &http.Client{
		Transport: &http.Transport{
			Dial:                dialer.Dial,
			TLSHandshakeTimeout: dialTimeout,
			MaxIdleConnsPerHost: maxIdle,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: fullTimeout,
	}
	return client, dialer
}

type Client struct {
	ApiKey    string
	SecretKey string
	URL       string
}

func (c *Client) buildURL(command string, params map[string]string) (string, error) {
	params["command"] = command
	params["response"] = "json"
	params["apiKey"] = c.ApiKey
	var sortedKeys []string
	for k := range params {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	var stringParams []string
	for _, key := range sortedKeys {
		queryStringParam := fmt.Sprintf("%s=%s", key, url.QueryEscape(params[key]))
		stringParams = append(stringParams, queryStringParam)
	}
	queryString := strings.Join(stringParams, "&")
	digest := hmac.New(sha1.New, []byte(c.SecretKey))
	digest.Write([]byte(strings.ToLower(queryString)))
	signature := base64.StdEncoding.EncodeToString(digest.Sum(nil))
	return fmt.Sprintf("%s?%s&signature=%s", c.URL, queryString, url.QueryEscape(signature)), nil
}

func (c *Client) Do(cmd string, params map[string]string, result interface{}) error {
	u, err := c.buildURL(cmd, params)
	if err != nil {
		return err
	}
	client := dial5Full300ClientNoKeepAlive
	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected response code for %s command %d: %s", cmd, resp.StatusCode, string(body))
	}
	if result != nil {
		err = json.Unmarshal(body, result)
		if err != nil {
			return fmt.Errorf("Unexpected result data for %s command: %s - Body: %s", cmd, err.Error(), string(body))
		}
	}
	return nil
}
