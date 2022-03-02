package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sort"
)

var wsChan = make(chan WsPayload)

var clients = make(map[WebSocketConnection]string)

var views = jet.NewSet(jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode())

var upgradeConn = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSJson struct {
	Action      string `json:"action"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WebSocketConnection struct {
	*websocket.Conn
}

type WsPayload struct {
	Action   string              `json:"action"`
	Message  string              `json:"message"`
	UserName string              `json:"username"`
	Conn     WebSocketConnection `json:"-"`
}

func WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConn.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("cient connected to endpoint")
	var response WSJson
	response.Message = `<em>connected</em>`
	conn := WebSocketConnection{ Conn:ws}
	clients[conn] = ""
	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}
	go ListenForWas(&conn)

}

func ListenForWas(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payLoad WsPayload
	for {
		err := conn.ReadJSON(&payLoad)
		if err != nil {

		} else {
			payLoad.Conn = *conn
			wsChan <- payLoad
		}
	}
}

func ListenToWsChannel() {
	var response WSJson
	for {
		e := <-wsChan
		switch e.Action {
		case "username":
			clients[e.Conn]=e.UserName
			users := GetUserList()
            response.Action = "list_users"
			response.ConnectedUsers = users
			BroadCast(response)
		case "left":
			delete(clients, e.Conn)
			users := GetUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			BroadCast(response)
		case "broadcast":
			response.Action="broadcast"
			response.Message=fmt.Sprintf("<strong>%s</strong>: %s", e.UserName, e.Message)
			BroadCast(response)


		}
		//response.Action = "Got Here"
		//response.Message = fmt.Sprintf("Some Message and action was %s ", e.Action)

	}
}

func GetUserList() []string {
	var users []string
	for _, x := range clients {
		if x != "" {
			users= append(users, x)
		}

	}
	sort.Strings(users)
	return users
}

func BroadCast(response WSJson) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket err")
			_ = client.Close()
			delete(clients, client)

		}
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

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
