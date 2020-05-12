package mutant_map_annotator

import (
	"crypto/rand"
	"fmt"
)

func UniqueShortId() (uuid string) {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%x", b[0:])
	return
}
