package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_"os"
	_"os/exec"
	_"runtime/pprof"
	"strconv"
	_"syscall"

	_"github.com/creack/pty"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	
)
type AllSessionManager struct{
	total_sessions int
	AllSessionMap map[int]*Session

}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (AllSessionManagerLocalObj *AllSessionManager) New_Session(w http.ResponseWriter, r *http.Request) {
    // for now i am just creating the  modifiy into websockete connection 
	// and then creating a new session and adding into AllSessionMap the new session 
	// but before doing this i have to check proper connection is setup or not 
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to connect the new session connection with terminal:", err)
		return
	}

	// also i have to handle url for logging clients to session and for now i will be using Session id
	// add lock here
    AllSessionManagerLocalObj.total_sessions+=1;
	// unlock here // i think this whole function should be in lock
    NewSessionObject:=IntializeSession(ws,AllSessionManagerLocalObj.total_sessions)
	AllSessionManagerLocalObj.AllSessionMap[AllSessionManagerLocalObj.total_sessions]=NewSessionObject



}
func (AllSessionManagerLocalObj *AllSessionManager) AddNewClient (w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Session_id string
	}
	 decoder := json.NewDecoder(r.Body)
	 var params  parameters
	 err := decoder.Decode(&params)
	fmt.Println(params)
	if err != nil {
		// respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		fmt.Println("error in decoding the request", err)
		return 
		// handle here properly

	} 
    strVar:=params.Session_id
	intVar, err := strconv.Atoi(strVar)
	if err != nil{
		fmt.Println(err)
		return 
	}
	session_obj:=AllSessionManagerLocalObj.AllSessionMap[intVar]
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade http connection of new client into websocket connection", err)
		return
	}
	session_obj.SessionMainChannel<-ws

    
  
}
func main(){
	myMap := make(map[int]*Session)

	// Create an instance of AllSessionManager and initialize the map
	Manager_obj := &AllSessionManager{
		AllSessionMap: myMap, 
		total_sessions :0,
	}
	router := chi.NewRouter()


	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		//AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()
    router.Mount("/v1", v1Router)
	srv := &http.Server{
		Addr:    ":" + "8088",
		Handler: router,
	}

    v1Router.HandleFunc("/NewSession", Manager_obj.New_Session)
	v1Router.HandleFunc("/NewClient",Manager_obj.AddNewClient)

	fmt.Println("server is listening on 8080")
	//log.Fatal(srv.ListenAndServe())
	//fmt.Println(("servers is staerd"))
    

	var err1 = srv.ListenAndServe()
	if err1 != nil {
		log.Fatal(err1)
	}
	fmt.Println(("servers is staerd"))



}