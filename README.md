# TTY_SERVER
<!-- code for new client is remaining that is to broadcast it about to all other clients and also to add in the database with new sequence number and to manage the type+payload that is requried to be sent is remaining -->
session will start with 1 sequence number


I have to properly handle the chunks because my result that needs to be sent to all clients will be large so it needs to sent in chunks and it should look like a terminal coutput at the client sides

code for synchronizing the info about current client is remaining

code for generating url for session at start of session is remaining and sending that url to terminal the session is remaining

code for adding redis remaining

code for removing client is remaining
 
code for closing session , clients for that session and closeing all goroutines associted with that thread is remaining and cleaning or erasing memory

~~code for connecting with  terminal server , clients frontend , and this TTY_Server is remaining~~

~~so i am able to connect terminal, client frontend, tty server locally now i have to open 3,4 clients on browser~~

sudo command is not working so i have to handle that thing

managing giving unique id to client as well as the new rooms

command for running client frontend : python3 -m http.server 8000

i think with url to join the room , the url can contain the room id so when request comes from a client ,it will also contain the room id , so in central server go we can write a function to parse request or get that room id

code for the ping pong is remaining which is to check wether the client is connected or not  // or for now we can handle at time of sending or receiving any message from websocket connection so whenever this type of case happens a notify will triger new event of removing this client from the all dashboard and also removing from all datastructures of the backend.


1) add redis functionality using container to work as database  , this also involves for managing seession and logic for clients 
2) then decide and implement logic for removing or crunching data from redis to avoid memeory full problem
3) add code to remove client or close client 
4) similarly for removing sessions or rooms
5) or add the functionality to work with url
6) manage properly necessary data and correct formatted get send to client and at frontend client or arranges data properly

