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
// sends the list of connected users when required
type WsJSONResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
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

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	log.Println("Client connected to endpoint")

	res := new(WsJSONResponse)
	res.Message = `<span style="color: cyan; font-style: italic;">Connected!</span>`
	res.ConnectedUsers = getOnlineUsers()
	res.Action = "list_users"
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
			// log.Println(err)
			//	do nothing
			continue
		}
		payload.Conn = *conn
		wsChan <- *payload
	}
}

// goroutine to listen for any new payload and then broadcast it to all clients
// based on "username" action, it adds associates a username with a connection
// based on "left" action, it deletes the connection
func ListenToWsChannel() {
	response := new(WsJSONResponse)
	var currClient string
	for {
		e := <-wsChan

		switch e.Action {
		case "username":
			currClient = clients[e.Conn]
			clients[e.Conn] = e.Username
			response.Action = "list_users"
			if currClient == "" {
				response.Message = fmt.Sprintf(`%s <span style="color: green; font-weight: bold;">joined!</span>`, e.Username)
			} else {
				response.Message = fmt.Sprintf(`%s<span style="color: cyan; font-weight: bold;"> is now </span>%s!`, currClient, e.Username)
			}

			response.ConnectedUsers = getOnlineUsers()

		case "left":
			if clients[e.Conn] != "" {
				response.Message = fmt.Sprintf(`%s <span style="color: red; font-weight: bold;">left!</span>`, clients[e.Conn])
			}
			response.Action = "list_users"
			delete(clients, e.Conn)
			response.ConnectedUsers = getOnlineUsers()
		default:
			log.Println("Incorrect Action")
			response.Action = "Got Here"
			response.Message = fmt.Sprintf("Some message for the action: %s", e.Action)
		}

		broadcastToAll(response)
	}
}

// broadcasts the response to all clients
// removes a client if there is any error in broadcasting the message
func broadcastToAll(response *WsJSONResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Printf("websocket error for %s\n", clients[client])
			_ = client.Close()
			delete(clients, client)
		}
	}
}

// gets the list of currently active users
func getOnlineUsers() []string {
	onlineUsers := []string{}
	for _, onlineUser := range clients {
		if onlineUser != "" {
			onlineUsers = append(onlineUsers, onlineUser)
		}
	}

	return onlineUsers
}

// function to render any page using `ResponseWriter` to write the response,
// `tmpl` to get the template name inside `views`
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
