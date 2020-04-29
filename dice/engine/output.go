package engine

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

func SendResponse(out *websocket.Conn, response []byte) error {
	log.Printf("%s\n", response)
	err := out.WriteMessage(websocket.TextMessage, response)
	if err != nil {
		log.Printf("write error: %s\n", err)
	}
	return err

}

func SendResponsef(out *websocket.Conn, format string, v ...interface{}) error {
	log.Printf(format, v...)
	err := out.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(format, v...)))
	if err != nil {
		log.Printf("write error: %s\n", err)
	}
	return err

}
