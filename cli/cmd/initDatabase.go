// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/notbaab/mutant-map-annotator"
	"github.com/spf13/cobra"
	"io/ioutil"
)

// query drops the database and adds test data
var deleteDatabase = `
DROP TABLE Game;
`
var populateDatabase = `
INSERT INTO Game (url, game_data) VALUES('12345', "{}");
`

// initDatabaseCmd represents the initDatabase command
var initDatabaseCmd = &cobra.Command{
	Use:   "initDatabase",
	Short: "Put test data into the database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		room, _ := cmd.Flags().GetString("room")
		database, _ := cmd.Flags().GetString("database")
		if room == "" {
			fmt.Println("No Room")
			// don't do anything
			return
		}

		if database == "" {
			fmt.Println("No database")
			// don't do anything
			return
		}

		fmt.Println("initDatabase called")
		// I'm doing a bad job here but eh
		db := mutant_map_annotator.SetupDatabase(database)
		mutant_map_annotator.RunStatment(db, deleteDatabase)
		db = mutant_map_annotator.SetupDatabase(database)

		mutant_map_annotator.RunStatment(db, populateDatabase)
		game, err := mutant_map_annotator.GetGame(db, 1)
		if err != nil {
			fmt.Println("Error", err.Error())
			return
		}

		fmt.Printf("%+v", game)

		if room == "" {
			fmt.Println("No Room")
			// don't do anything
			return
		}
		dat, err := ioutil.ReadFile(room)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		game.GameData = string(dat)
		mutant_map_annotator.UpdateGame(db, game)
		fmt.Printf("%+v", game)
	},
}

func init() {
	rootCmd.AddCommand(initDatabaseCmd)
	initDatabaseCmd.Flags().StringP("room", "r", "", "Room file to populate game with")
	initDatabaseCmd.Flags().StringP("database", "d", "", "da database")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initDatabaseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initDatabaseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
