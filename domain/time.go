package domain

import (
	"github.com/ch629/go-irc-kafka/irc/parser"
	"strconv"
	"time"
)

func timeFromTmiSentTs(tags parser.Tags) (time.Time, error) {
	ts, err := strconv.Atoi(tags["tmi-sent-ts"])
	if err != nil {
		return time.Time{}, err
	}
	t := time.Unix(0, int64(ts*int(time.Millisecond)))
	return t, nil
}
