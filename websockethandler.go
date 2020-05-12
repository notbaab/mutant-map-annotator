// backend node for managing connections
package mutant_map_annotator

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// a backend controller abstracts handling and managing websocket connections
type NetworkController struct {
	db      string
	db_conn *sqlx.DB
	Rooms   map[string]*Room
}

func NewNetworkController(db string) NetworkController {
	controller := NetworkController{db: db, Rooms: make(map[string]*Room)}
	controller.db_conn = sqlx.MustConnect("sqlite3", db)
	return controller
}

// entry point for the web socket handler. Upgrades the websocket and starts the
// worker that handlers messages
func (nc NetworkController) WsHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	conn, err := websocket.Upgrade(writer, request, nil, 1024, 1024)

	roomId := vars["roomId"]

	Info.Println("Doing a websocket")

	if _, ok := err.(websocket.HandshakeError); ok {
		log.Println("error")
		http.Error(writer, "got a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}

	// get or create the room.
	// TODO: Database query here
	room, exists := nc.Rooms[roomId]
	if !exists {
		room, err = NewRoom(roomId, nc.db)
		if err != nil {
			Error.Printf("Couldn't create new room %s, error %s", roomId, err.Error())
			return
		}
		Error.Printf("%+v", room)
		nc.Rooms[roomId] = room
		go room.run()
	}

	// Initialize a client object in the correct room
	client, err := room.AddClient(conn)
	if err != nil {
		Error.Printf("Something happened %s", err.Error())
		return
	}

	Info.Printf("Added client %s to room %s", client.Id, roomId)
}

func (nc NetworkController) FetchRoomState(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	roomId := vars["roomId"]

	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	game, err := FindGame(nc.db_conn, roomId)
	if err != nil {
		Error.Printf("no game found for %s, error %s", roomId, err.Error())
		errMsg := ErrorMessage{Error: err.Error()}
		jsonStr, _ := json.Marshal(errMsg) // if we can't marshel the error message we should give up on life
		writer.Write(jsonStr)
		return
	}

	mapJson, err := json.Marshal(game.GameData)
	if err != nil {
		Error.Printf("Couldn't marshal game data for room %s, error %s", roomId, err.Error())
		errMsg := ErrorMessage{Error: err.Error()}
		jsonStr, _ := json.Marshal(errMsg) // if we can't marshel the error message we should give up on life
		writer.Write(jsonStr)
		return
	}

	// rawJsonData := json.RawMessage(mapJson)
	// stateMessage := Message{Event: "state", Data: &rawJsonData}
	// response, err := json.Marshal(stateMessage)

	// if err != nil {
	// 	Error.Printf("Couldn't marshal game data for room %s, error %s", roomId, err.Error())
	// 	errMsg := ErrorMessage{Error: err.Error()}
	// 	jsonStr, _ := json.Marshal(errMsg) // if we can't marshel the error message we should give up on life
	// 	writer.Write(jsonStr)
	// 	return
	// }

	Info.Printf("Writing %s", mapJson)

	writer.Write(mapJson)
}

// Some leftover code that has the server establishing a websocket connection with
// a url. It's useful and I don't have the heart to remove it yet
// func (b NetworkController) NewWebsocket(connectionUrl string) (*websocket.Conn, error) {
// 	u, err := url.Parse(connectionUrl)
// 	if err != nil {
// 		log.Println(err)
// 		return nil, errors.New("Cannot parse connection url" + connectionUrl)
// 	}
// 	log.Println(u)

// 	log.Println(u.Host)
// 	rawConn, err := net.Dial("tcp", u.Host)
// 	if err != nil {
// 		log.Println(err)
// 		return nil, errors.New("cannot dial " + u.Host)
// 	}

// 	wsHeaders := http.Header{
// 		"Origin": {u.Host},
// 		// your milage may differ
// 		"Sec-WebSocket-Extensions": {
// 			"permessage-deflate; client_max_window_bits, x-webkit-deflate-frame"},
// 	}

// 	wsConn, resp, err := websocket.NewClient(rawConn, u, wsHeaders, 1024, 1024)

// 	if err != nil {
// 		return nil, fmt.Errorf("websocket.NewClient Error: %s\nResp:%+v", err, resp)
// 	}

// 	b.AddNewConnection(wsConn)
// 	return wsConn, nil
// }
