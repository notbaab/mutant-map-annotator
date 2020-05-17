package mutant_map_annotator

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
)

type State interface {
	BuildAddMessage() Message
}

// keeps track of data from a client
// Ignores Input queue and Socket when marshalling
type Client struct {
	Socket   *websocket.Conn `json:"-"`
	Room     *Room
	sendChan chan Message

	Role string
	Id   string
}

type ClientMessage struct {
	Id   string `json:"id"`
	Data json.RawMessage
}

func NewClient(conn *websocket.Conn, room *Room) (client Client) {
	newClient := Client{}

	newClient.Socket = conn
	newClient.Room = room
	newClient.sendChan = make(chan Message, 5)
	// TODO: Do we trust unique short id to be unique? Yeah probably.
	newClient.Id = UniqueShortId()

	return newClient
}

func (c *Client) writeRoutine() {
	defer func() {
		c.Room.unregister <- c
		c.Socket.Close()
	}()
	for {
		// do with a select, we will probably want more channels to communicate
		// with the client at some point
		select {
		case message, ok := <-c.sendChan:
			c.Socket.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.send(message)
		}
	}
}

func (c *Client) readRoutine() {
	defer func() {
		c.Room.unregister <- c
		c.Socket.Close()
	}()

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				Error.Printf("error: %v", err)
			}
			break
		}

		err = c.HandleMessage(message)
		if err != nil {
			Error.Printf("error: %v", err)
		}
	}
}

func (c *Client) HandleMessage(event []byte) error {
	var message Message
	err := json.Unmarshal(event, &message)
	if err != nil {
		Error.Printf("Got error %s", err.Error())
		return err
	}
	return c.Room.doMessage(message, c)
}

func (c *Client) send(msg Message) error {
	Trace.Printf("Sending something")
	rawMessage, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	Trace.Printf("Sending %s", string(rawMessage))

	if err := c.Socket.WriteMessage(websocket.TextMessage, rawMessage); err != nil {
		// c.Room.CleanUpHandler(c)
		return err
	}

	return nil
}
