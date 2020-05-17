package mutant_map_annotator

import (
	// "encoding/json"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Room struct {
	Clients map[*Client]bool

	game Game

	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	db         string
}

func (r *Room) CleanUpHandler(c Client) {
	Info.Printf("cleaning up connection from %s\n", c.Socket.RemoteAddr())
	err := c.Socket.Close()

	if err != nil {
		Error.Println(err.Error())
	}
}

func NewRoom(url, database string) (*Room, error) {
	Info.Printf("Data base is %s", database)
	db := sqlx.MustConnect("sqlite3", database)
	game, err := FindGame(db, url)

	if err != nil {
		return nil, err
	}

	return &Room{
		Clients:    make(map[*Client]bool),
		broadcast:  make(chan Message, 5),
		register:   make(chan *Client, 5),
		unregister: make(chan *Client, 5),
		game:       game,
		db:         database,
	}, nil
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.register:
			r.Clients[client] = true
		case client := <-r.unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.sendChan)
			}
		case message := <-r.broadcast:
			for client := range r.Clients {
				select {
				case client.sendChan <- message:
				default:
					close(client.sendChan)
					delete(r.Clients, client)
				}
			}
		}
	}
}

// func CreateMessage()

func (r *Room) doMessage(message Message, srcClient *Client) error {
	Trace.Printf("Got a thing %+v data is %s\n", message, string(*message.Data))
	if message.Event == "state" {
		// Should I reconnect everytime? Maybe? Probably not
		db := sqlx.MustConnect("sqlite3", srcClient.Room.db)
		id := srcClient.Room.game.Id
		game, err := GetGame(db, id)

		if err != nil {
			Error.Printf("No game found for game id %d, how'd you get here?", id)
			return err
		}

		game.GameData = string(*message.Data)
		UpdateGame(db, game)

		if err != nil {
			Error.Printf("No game found for game id %d, how'd you get here?", id)
			return err
		}

		fullMessage := Message{Event: "state", Data: message.Data}
		srcClient.send(fullMessage)

		for c, _ := range r.Clients {
			c.send(fullMessage)
		}
	}

	return nil
}

func (r *Room) AddClient(conn *websocket.Conn) (Client, error) {
	client := NewClient(conn, r)

	r.register <- &client

	go client.readRoutine()
	go client.writeRoutine()

	return client, nil
}
