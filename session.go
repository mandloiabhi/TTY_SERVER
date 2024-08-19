package main

import (
	"fmt"
	"log"
	_ "math/rand"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type Session struct {
	SessionId    int
	AllClientMap map[int]*Client
	// room url
	// Redis in memory database
	MemoryDatabase     map[int]string
	SequenceNo         int
	Terminal_conn      *websocket.Conn
	SessionMainChannel chan *websocket.Conn
	Terminal_Channel   chan []byte
	Synch_Channel      chan []byte // carefull here type+data both will be there , this is to be handle
}

func (session *Session) ReceiveNewClient() {
	Client_Count := 0
	for {

		new_client_conn := <-session.SessionMainChannel
		NewClientObj := session.New_Client(new_client_conn, Client_Count)
		go NewClientObj.StartClient()
		NewClientObj.addnewclient()
		session.AllClientMap[Client_Count] = NewClientObj
		Client_Count += 1
		fmt.Println("client added")
		fmt.Println(Client_Count)

	}
}

func (session *Session) TerminalRead() {
	for {
		messageType, payload, err := session.Terminal_conn.ReadMessage()
		// for now i am not handling different type of messageType // i am considering that myTerminal is always running
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}
		log.Println("MessageType: ", messageType)
		log.Println("Payload: ", string(payload))
		// 1 for terminal output data
		messageStr := fmt.Sprintf("%d^%s", 1, payload)
		combinedMessage := []byte(messageStr)
		session.Synch_Channel <- combinedMessage // this has to be handle properly

		// now instead of printing i have to save this message first in the session database in Redis along with new sequence number for this chunk
		//TO  DO: maintain proper chunk carefully take care of size , take care of data type in payload
		// TO DO : After saving I have to intiate Notiy_all so to broadcast this changes to all clients
		// TO DO : I HAVE TO pass this message to the some gorouting using channel Terminal_Channel

	}
}
func (session *Session) TerminalWrite() {
	for {
		select {
		// this will received when we have already saved the request into the database and all ready broadcasted to all the clients
		case message, ok := <-session.Terminal_Channel:
			// Ok will be false Incase the egress channel is closed
			if !ok {
				// Manager has closed this connection channel, so communicate that to frontend
				if err := session.Terminal_conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					// Log that the connection is closed and the reason
					log.Println("connection closed: ", err)
				}
				// Return to close the goroutine
				return
			}
			// Write a Regular text message to the connection
			if err := session.Terminal_conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println(err)
			}
			log.Println("sent command to channel")
		}
	}
}
func (session *Session) MaintainSession() {

	// here i have to get info from the Synch channel
	// take action based on the type of info
	// if it is command comming from client then firstly store in database as new seqeuence number  then second broadcast it to all clients through their Client_Channel  then pass this to Terminal_channel to run the command
	// if it is output comming from terminal through Synch channel store in database as new sequence then broadcast to all client again through Client_Channel

	// so we are receiving any event info from Synch channel and broadcasting or sending to another channel

	// and all of this done in for loop which continiously runing and take care of locks , mutex requrement

	for {
		receivedMessage := <-(session.Synch_Channel)

		// Split the received message into message type and payload.
		msg := fmt.Sprintf("%s", receivedMessage)
		parts := strings.SplitN(string(receivedMessage), "^", 2)
		if len(parts) != 2 {
			log.Println("Invalid message format")
			return
		}

		// Convert the message type back to an int.
		receivedMessageType, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("Invalid message type: %v", err)
			return
		}

		// The payload is the second part.

		//receivedPayload := []byte(parts[1])
		session.SequenceNo = session.SequenceNo + 1
		// saved in database along with sequence number
		session.MemoryDatabase[session.SequenceNo] = msg

		session.Notify_all()
		// i need here asynchronous thing
		if receivedMessageType == 2 { // this is for giving command to terminal to run it
			receivedPayload := []byte(parts[1])
			session.Terminal_Channel <- receivedPayload
		}

	}

}
func (session *Session) Notify_all() {

	st := "pass"
	msg := []byte(st)
	for key, value := range session.AllClientMap {
		value.Client_Channel <- msg
		fmt.Println(key)
	}

}
func IntializeSession(conn_with_terminal *websocket.Conn, Id int) *Session {
	myMap := make(map[int]*Client)
	databasemap := make(map[int]string)
	Session_obj := &Session{
		SequenceNo:         0,
		MemoryDatabase:     databasemap,
		AllClientMap:       myMap,
		SessionId:          Id,
		Terminal_conn:      conn_with_terminal,
		SessionMainChannel: make(chan *websocket.Conn),
		Terminal_Channel:   make(chan []byte),
		Synch_Channel:      make(chan []byte),
	}
	go Session_obj.TerminalRead()
	go Session_obj.TerminalWrite()
	go Session_obj.ReceiveNewClient()
	go Session_obj.MaintainSession()

	return Session_obj
}
