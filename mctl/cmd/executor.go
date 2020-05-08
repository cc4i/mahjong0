package cmd

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
)

var apiVersion = "v1alpha1"

func RunPost(addr string, uri string, body []byte) (int, error) {
	u := &url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   fmt.Sprintf("/%s/%s", apiVersion, uri),
	}
	resp, err := http.Post(u.String(), "text/yaml", bytes.NewReader(body))
	if err != nil {
		log.Printf("%s\n", err)
	} else {
		log.Printf("%s\n", resp.Body)
	}
	return resp.StatusCode, err
}

func Run(addr string, dryRun bool, cmd []byte) error {
	if c, err := Connect2Dice(addr, dryRun); err != nil {
		log.Printf("failed to connect with Dice: %s \n", err)
		return err
	} else {
		return ExecCommand(cmd, c)
	}
}

func Connect2Dice(addr string, dryRun bool) (*websocket.Conn, error) {
	u := &url.URL{
		Scheme: "ws",
		Host:   addr,
		Path:   fmt.Sprintf("/%s/%s", apiVersion, "ws"),
	}
	if dryRun {
		u.Path = fmt.Sprintf("/%s/%s", apiVersion, "ws?dryRun=true")
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return c, err
	}
	return c, nil
}

func ExecCommand(cmd []byte, c *websocket.Conn) error {
	if err := c.WriteMessage(websocket.TextMessage, cmd); err != nil {
		log.Printf("write error: %s\n", err)
		return err
	}
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("read error: %s\n", err)
			return err
		}
		if string(message) == "d-done" {
			return nil
		} else {
			log.Printf("%s\n", message)
		}
	}
}
