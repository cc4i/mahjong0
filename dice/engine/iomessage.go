package engine

import (
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

var mutex sync.Mutex

// SR send repose back to client & output logs
func SR(out *websocket.Conn, response []byte) {
	log.Printf("%s\n", response)
	mutex.Lock()
	defer mutex.Unlock()
	err := out.WriteMessage(websocket.TextMessage, response)
	if err != nil {
		log.Printf("write error: %s\n", err)
	}

}

// SRf send repose back to client & output logs with given format
func SRf(out *websocket.Conn, format string, v ...interface{}) {
	log.Printf(format, v...)
	mutex.Lock()
	defer mutex.Unlock()
	err := out.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(format, v...)))
	if err != nil {
		log.Printf("write error: %s\n", err)
	}

}

// SRF send repose back to client, output logs & files
func SRF(out *websocket.Conn, file *os.File, response []byte) {
	log.Printf("%s\n", response)
	mutex.Lock()
	defer mutex.Unlock()
	err := out.WriteMessage(websocket.TextMessage, response)
	if err != nil {
		log.Printf("write error: %s\n", err)
	}
	_, err = file.Write(response)
	if err != nil {
		log.Printf("write error: %s\n", err)
	}
}
