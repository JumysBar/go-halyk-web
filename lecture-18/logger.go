package main

import (
	"errors"
	"github.com/rs/zerolog/log"
)

var _err = errors.New("some error")

func main() {
	var _log = struct {
		Message string `json:"message"`
		Err     error  `json:"err"`
	}{
		"информация об объекте над которым происходит что-то",
		_err,
	}

	//fmt.Printf("%s: %s, %s\n", "error", _log.Err, _log.Message)
	//
	log.Printf("%s: %s, %s\n", "error", _log.Err, _log.Message)

	log.Error().Err(_log.Err).Msgf("%s", _log.Message)

}
