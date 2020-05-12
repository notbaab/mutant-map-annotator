package mutant_map_annotator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

type ZoneRow struct {
	CoordinateStr  string
	X              int
	Y              int
	Explored       bool
	ArtifactExists bool
	ArtifactLooted bool
	ArtifactType   string
	ScrapLooted    bool
	ScrapeType     string
	Tower          bool
	Environmnet    string
	RotLevel       int
	ThreatType     string

	ZoneNotes string

	FTPArtifact    bool
	FTPRotLevel    bool
	FTPBullets     bool
	FTPRations     bool
	FTPWater       bool
	FTPExploration bool
	FTPRotFinder   bool
	FTPRush        bool
}

func (zr *ZoneRow) loadRowStr(row []interface{}) error {
	// Todo Make this configurable
	index := 0
	zr.CoordinateStr = row[index].(string)
	index = index + 1
	zr.Explored = row[index].(string) == "TRUE"
	index = index + 1
	zr.ArtifactExists = row[index].(string) == "TRUE"
	index = index + 1
	zr.ArtifactLooted = row[index].(string) == "TRUE"
	index = index + 1
	zr.ArtifactType = row[index].(string)
	index = index + 1
	zr.ScrapeType = row[index].(string)
	index = index + 1
	zr.ScrapeType = row[index].(string)
	index = index + 1
	zr.Tower = row[index].(string) == "TRUE"
	index = index + 1
	zr.Environmnet = row[index].(string)
	index = index + 1

	rot, err := strconv.Atoi(row[index].(string))
	index = index + 1
	if err != nil {
		// fmt.Printf("fuck '%s'\n", row[8])
		rot = 0
	}

	zr.RotLevel = rot

	zr.ThreatType = row[index].(string)
	index = index + 1
	zr.ZoneNotes = row[index].(string)
	index = index + 1
	zr.FTPArtifact = row[index].(string) == "TRUE"
	index = index + 1
	zr.FTPRotLevel = row[index].(string) == "TRUE"
	index = index + 1
	zr.FTPBullets = row[index].(string) == "TRUE"
	index = index + 1
	zr.FTPRations = row[index].(string) == "TRUE"
	index = index + 1
	zr.FTPWater = row[index].(string) == "TRUE"
	index = index + 1
	zr.FTPExploration = row[index].(string) == "TRUE"
	index = index + 1
	zr.FTPRotFinder = row[index].(string) == "TRUE"
	index = index + 1
	zr.FTPRush = row[index].(string) == "TRUE"
	index = index + 1

	yAsLetter := zr.CoordinateStr[0]
	zr.Y = charToCoordinate(yAsLetter)
	zr.X, err = strconv.Atoi(zr.CoordinateStr[1:])
	if err != nil {
		fmt.Printf("fuck, not a number in the coordinate str '%+v'\n", row)
		return errors.New("at the end most likely")
	}
	// 0 based grid cause we aren't a monster
	zr.X = zr.X - 1
	return nil
}

type ZoneSheet struct {
	Grid [][]ZoneRow `json:"grid"`
}

// Expects the byte to be an upper case value
func charToCoordinate(c byte) int {
	return int(c - 'A')

}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func loadSheet() (ZoneSheet, error) {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		return ZoneSheet{}, err
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	spreadsheetId := "1M8VegAWfFK5oUDnKPbamNuZdms5ENBvQU3Dv8ZJWa2U"
	readRange := "Zone Info!A2:T"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
		return ZoneSheet{}, errors.New("No data found")
	}

	sheet := ZoneSheet{}
	maxX := 0
	maxY := 0

	rows := make([]ZoneRow, 100)
	for _, row := range resp.Values {
		zoneRow := ZoneRow{}
		err := zoneRow.loadRowStr(row)
		if err != nil {
			continue
		}

		if zoneRow.X > maxX {
			maxX = zoneRow.X
		}
		if zoneRow.Y > maxY {
			maxY = zoneRow.Y
		}
		rows = append(rows, zoneRow)
	}
	maxX = maxX + 1
	maxY = maxY + 1

	fmt.Printf("%d and %d\n", maxX, maxY)

	sheet.Grid = make([][]ZoneRow, maxX)
	for x := 0; x < maxX; x++ {
		sheet.Grid[x] = make([]ZoneRow, maxY)
	}

	for _, zone := range rows {
		// fmt.Printf("%+v\n", rows)
		x := zone.X
		y := zone.Y

		sheet.Grid[x][y] = zone

		// sheet.Rows = append(sheet.Rows, zone)
		// Print columns A and E, which correspond to indices 0 and 4.
		// fmt.Printf("%s ", zone.CoordinateStr)
		// if zone.Explored {
		// 	fmt.Printf("%d,%d  %b\n", zone.X, zone.Y, zone.Explored)
		// }
		// fmt.Printf("\n")
	}
	return sheet, nil
}
