package main

import (
	"github.com/gorilla/websocket"
	"log"
)

type Session struct {
	SessionId    int
	AllClientMap map[int]*Client
	// room url
	// Redis in memory database
	Terminal_conn *websocket.Conn
	SessionMainChannel chan *websocket.Conn
	Terminal_Channel chan []byte
	Synch_Channel chan []byte // carefull here type+data both will be there , this is to be handle
}
func (session *Session) ReceiveNewClient() {
	Client_Count:=0
	for {
		new_client_conn := <-session.SessionMainChannel
        NewClientObj:=session.New_Client(new_client_conn,Client_Count)
		NewClientObj.StartClient()
		session.AllClientMap[Client_Count]=NewClientObj
		Client_Count+=1
		


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

			
			//session.Synch_Channel<-payload   // this has to be handle properly


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
	

}

func IntializeSession(conn_with_terminal *websocket.Conn, Id int) *Session {
	myMap := make(map[int]*Client)
	Session_obj := &Session{
		AllClientMap: myMap,
		SessionId:     Id,
		Terminal_conn: conn_with_terminal,
		SessionMainChannel: make(chan *websocket.Conn),
		Terminal_Channel: make(chan []byte),
	}
	go Session_obj.TerminalRead()
	go Session_obj.TerminalWrite()
	go Session_obj.ReceiveNewClient()
	go Session_obj.MaintainSession()


	return Session_obj
}
