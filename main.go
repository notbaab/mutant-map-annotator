package mutant_map_annotator

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	// "github.com/jmoiron/sqlx"
)

func cleanUpSocket(conn *websocket.Conn) *Message {
	Info.Printf("Cleaning up connection from %s\n", conn.RemoteAddr())
	err := conn.Close()

	if err != nil {
		Error.Println(err.Error())
	}
	return nil
}

// TODO: SHould this be in server vars?
func setupFileServer(dir string) http.Handler {
	// handle all requests by serving a file of the same name
	Info.Printf("Starting with %s", dir)
	fs := http.Dir(dir)
	fileHandler := http.FileServer(fs)
	return fileHandler
}

func setupRoutes(staticFolder string, controller NetworkController) *mux.Router {
	r := mux.NewRouter()
	wapi := r.PathPrefix("/ws").Subrouter()
	wapi.HandleFunc("/{roomId}", controller.WsHandler)

	roomRestEndpoints := r.PathPrefix("/room-data").Subrouter()
	roomRestEndpoints.HandleFunc("/{roomId}", controller.FetchRoomState)

	// TODO: The static router should be done in nginx I think
	staticFileHandler := setupFileServer(staticFolder)
	r.PathPrefix("/").Handler(staticFileHandler)

	return r
}

func Start(frontend, db string) {
	InitLogger(os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	controller := NewNetworkController(db)
	mux := setupRoutes(frontend, controller)

	Info.Println("Starting")
	runHttpServer("0.0.0.0:5658", mux)
}

func runHttpServer(addr string, mux http.Handler) {
	// this call blocks -- the progam runs here forever
	err := http.ListenAndServe(addr, mux)
	Warning.Println(err.Error())
}
