package mutant_map_annotator

import (
	"io"
	"log"
)

// Global Vars Maybe make them just a package at one point?
var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func InitLogger(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"\033[0;36mTRACE: \033[0m",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"\033[0;32mINFO: \033[0m",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"\033[0;33mWARNING: \033[0m",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"\033[0;31mERROR: \033[0m",
		log.Ldate|log.Ltime|log.Lshortfile)
}
