package main

import (
	_ "encoding/json"
	"fmt"
	"log"
	_ "net/http"
	_ "os"
	_ "os/exec"
	_ "runtime/pprof"
	_ "strconv"
	"strings"
	_ "syscall"

	_ "github.com/creack/pty"
	_ "github.com/go-chi/chi/v5"
	_ "github.com/go-chi/cors"
	"github.com/gorilla/websocket"
)

type Client struct {
	client_id          int
	Client_conn        *websocket.Conn
	LastSequenceNumber int
	session            *Session
	Client_Channel     chan []byte // this channel is used for getting any output that needs to be sent to the client from the session manitainer goroutine

}

func (session *Session) New_Client(conn *websocket.Conn, ID int) *Client {

	Client_obj := &Client{
		Client_conn:        conn,
		client_id:          ID,
		LastSequenceNumber: 0,
		session:            session,
		Client_Channel:     make(chan []byte),
	}
	return Client_obj
}
func (client *Client) StartClient() {
	go client.ReadClient()
	go client.WriteClient()

}
func (client *Client) ReadClient() {
	for {
		messageType, payload, err := client.Client_conn.ReadMessage()
		// for now i am not handling different type of messageType // i am considering that myTerminal is always running
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}
		log.Println("MessageType: ", messageType)
		log.Println("Payload: ", string(payload))
		// 2 for the output for the command from client
		messageStr := fmt.Sprintf("%d^%s", 2, payload)
		combinedMessage := []byte(messageStr)
		fmt.Println("just before sending in synch channel")
		(client.session.Synch_Channel) <- combinedMessage
		fmt.Println("send to add to database and broadcast to all other clients")
		// client.session.Synch_Channel<-payload // this has to be handle properly as data type etc

		// basically messageType will be either command or some thing else
		// now instead of printing i have to save this message first in the session database in Redis along with new sequence number for this chunk
		//TO  DO: maintain proper chunk carefully take care of size , take care of data type in payload
		// TO DO : After saving I have to intiate Notiy_all so to broadcast this changes to all clients
		// TO DO : I HAVE TO pass this message to the some gorouting using channel Terminal_Channel

	}
}
func (client *Client) WriteClient() {
	for {
		select {
		// this will received when we have already saved the request into the database and all ready broadcasted to all the clients
		case message, ok := <-client.Client_Channel:
			// Ok will be false Incase the egress channel is closed
			if !ok {
				// Manager has closed this connection channel, so communicate that to frontend
				if err := client.Client_conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					// Log that the connection is closed and the reason
					log.Println("connection closed: ", err)
				}
				// Return to close the goroutine
				return
			}
			// Write a Regular text message to the connection
			var values []string
			values = client.Synchronize()
			combinedString := strings.Join(values, "")
			//fmt.Println(combinedString)
			//message_string := fmt.Sprintf("%d|%s", 5,combinedString)
			message = []byte(combinedString)
			if err := client.Client_conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println(err)
			}
			fmt.Println(combinedString)
			log.Println("sent result of command to client ")
		}
	}
}

func (client *Client) Synchronize() []string {

	givenInt := client.LastSequenceNumber

	// Slice to store values with keys greater than the given integer.
	var values []string

	// Iterate over the map.
	// i have to optimize it so it iterates from given seesion number
	for key, value := range client.session.MemoryDatabase {
		if key > givenInt {
			values = append(values, value)
		}
	}
	client.LastSequenceNumber = client.session.SequenceNo
	return values

}

func (client *Client) addnewclient() {
	st := "pass"
	msg := []byte(st)
	client.Client_Channel <- msg
}

// code for changing wether it is command or result of command is requried to be done
