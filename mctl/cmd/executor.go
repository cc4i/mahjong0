package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kris-nova/logger"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var apiVersion = "v1alpha1"

func RunGet(addr string, uri string) ([]byte, error) {

	resp, err := http.Get(fmt.Sprintf("http://%s/%s", addr, uri))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func RunGetByVersion(addr string, uri string) ([]byte, error) {

	uri = fmt.Sprintf("%s/%s", apiVersion, uri)
	return RunGet(addr, uri)
}

func RunPostByVersion(addr string, uri string, body []byte) (int, error) {
	u := &url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   fmt.Sprintf("/%s/%s", apiVersion, uri),
	}
	resp, err := http.Post(u.String(), "text/yaml", bytes.NewReader(body))
	if err != nil {
		logger.Warning("%s\n", err)
	} else {
		logger.Warning("%s\n", resp.Body)
	}
	return resp.StatusCode, err
}

func Run(addr string, dryRun bool, parallel bool, cmd []byte) error {
	if c, err := Connect2Dice(addr, dryRun, parallel); err != nil {
		logger.Warning("failed to connect with Dice: %s \n", err)
		return err
	} else {
		return ExecCommand(cmd, c)
	}
}

func Connect2Dice(addr string, dryRun bool, parallel bool) (*websocket.Conn, error) {
	u := &url.URL{
		Scheme: "ws",
		Host:   addr,
		Path:   fmt.Sprintf("/%s/%s", apiVersion, "ws"),
	}
	if dryRun {
		u.RawQuery = "dryRun=true"
		if parallel {
			u.RawQuery = "dryRun=true&parallel=true"
		}
	}
	if parallel {
		u.RawQuery = "parallel=true"
		if dryRun {
			u.RawQuery = "dryRun=true&parallel=true"
		}
	}
	logger.Info("Connecting to %s\n", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return c, err
	}
	return c, nil
}

func ExecCommand(cmd []byte, c *websocket.Conn) error {
	if err := c.WriteMessage(websocket.TextMessage, cmd); err != nil {
		logger.Warning("write error: %s\n", err)
		return err
	}
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			logger.Warning("read error: %s\n", err)
			return err
		}
		if string(message) == "d-done" {
			return nil
		} else {
			buf, _ := bufio.NewReader(bytes.NewReader(message)).ReadBytes('\n')
			str := strings.ToLower(string(buf))
			if strings.Contains(str, "warn") {
				logger.Warning("%s\n", buf)
			} else if strings.Contains(str, "error") || strings.Contains(str, " err ") {
				logger.Warning("%s\n", buf)
			} else {
				logger.Info("%s\n", buf)
			}

		}
	}
}
