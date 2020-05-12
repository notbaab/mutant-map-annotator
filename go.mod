module mutant-map-annotator

go 1.14

require (
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/jmoiron/sqlx v1.2.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/notbaab/mutant-map-annotator v0.0.0-00010101000000-000000000000 // indirect
	github.com/notbaab/mutant-map-annotator/cli/cmd v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.0.0 // indirect
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.22.0
)

replace github.com/notbaab/mutant-map-annotator/cli/cmd => ./cli/cmd

replace github.com/notbaab/mutant-map-annotator => ../mutant-map-annotator
