package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

// get the views from the folder containing the *.jet files
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

// used to upgrade the connection from HTTP to Websocket
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// channel used for communication between `ListenToWs` and `ListenToWsChannel`
// goroutines. payload is received from the first via the clients and processed
// by the second to send a particular response back.
var wsChan = make(chan WsPayload)

// keep track of the users
var clients = make(map[WebSocketConnection]string)

// type for response json that will be sent back to client
type WsJSONResponse struct {
	Action      string `json:"action"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
}

// type for a particular connection that is created by any client
type WebSocketConnection struct {
	*websocket.Conn
}

// type for message input from a particular client
type WsPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

// renders the Home page
func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

// used to create a new connection and then start listening on that connection
// upgrades the incoming http request to websocket and then sends back the
// confirmation to client. it starts a goroutine to listen for any messages
// from that particular client.
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client connected to endpoint")

	res := new(WsJSONResponse)
	res.Message = "Connected!"

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = "User"

	log.Println()

	err = ws.WriteJSON(res)
	if err != nil {
		log.Println(err)
	}

	go ListenForWs(&conn)
}

// goroutine to listen to any incoming messages from a particular `conn`
// connection. it sends the message to `wsChan` for further processing.
// has `recover()` to continue running even in case of any panic.
func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error - ", fmt.Sprintf("%+v", r))
		}
	}()

	payload := new(WsPayload)

	for {
		err := conn.ReadJSON(payload)
		if err != nil {
			log.Println(err)
			//	do nothing
		}
		payload.Conn = *conn
		wsChan <- *payload
	}
}

// goroutine to listen for any new payload and then broadcast it to all clients
func ListenToWsChannel() {
	response := new(WsJSONResponse)

	for {
		e := <-wsChan

		response.Action = "Got Here"
		response.Message = fmt.Sprintf("Some message for the action: %s", e.Action)

		broadcastToAll(response)
	}
}

// broadcasts the response to all clients
// removes a client if there is any error in broadcasting the message
func broadcastToAll(response *WsJSONResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Printf("websocket error for %s", clients[client])
			_ = client.Close()
			delete(clients, client)
		}
	}
}

// function to render any page using `ResponseWriter` to write the response, `tmpl` to get the template name inside `views`
// and `data` to show any data during render
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
