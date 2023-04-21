package main

import (
	"luraim/contact/contacts"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.DisableSampling(true)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	contacts.Run()
}
