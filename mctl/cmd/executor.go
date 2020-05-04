package cmd

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

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
		Path:   "/v1alpha1/ws",
	}
	if dryRun {
		u.Path = "/v1alpha1/ws?dryRun=true"
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
