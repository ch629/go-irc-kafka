package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Scan(t *testing.T) {
	reader := strings.NewReader("@test=abc;foo=bar :prefix cmd par1 par2: trailing \r\n")

	scanner := NewScanner(reader)

	msg, err := scanner.Scan()

	assert.NoError(t, err)
	assert.Equal(t, &Message{
		Tags: map[string]string{
			"test": "abc",
			"foo":  "bar",
		},
		Prefix:  "prefix",
		Command: "cmd",
		Params: []string{
			"par1",
			"par2",
			" trailing ",
		},
	}, msg)

}

func Test_Other(t *testing.T) {
	reader := strings.NewReader(":tmi.twitch.tv CAP * ACK :twitch.tv/tags twitch.tv/commands twitch.tv/membership\r\n")
	scanner := NewScanner(reader)

	msg, err := scanner.Scan()

	assert.NoError(t, err)

	assert.Equal(t, &Message{
		Prefix:  "tmi.twitch.tv",
		Command: "CAP",
		Params: []string{
			"*",
			"ACK",
			"twitch.tv/tags twitch.tv/commands twitch.tv/membership",
		},
	}, msg)
}

func Test_Other2(t *testing.T) {
	reader := strings.NewReader(":tmi.twitch.tv 001 thewolfpack :Welcome, GLHF!\r\n")
	scanner := NewScanner(reader)

	msg, err := scanner.Scan()

	assert.NoError(t, err)

	assert.Equal(t, &Message{
		Prefix:  "tmi.twitch.tv",
		Command: "001",
		Params: []string{
			"thewolfpack",
			"Welcome, GLHF!",
		},
	}, msg)
}

func Test_Other3(t *testing.T) {
	reader := strings.NewReader("@badge-info=subscriber/8;badges=subscriber/6,bits/75000;color=#1E90FF;display-name=Ovojaytee;emotes=1837404:44-50/915234:164-169/1093027:13-18;flags=;id=aa52e1d2-6ff5-42ba-b205-9d4a15f9dbf8;login=ovojaytee;mod=0;msg-id=resub;msg-param-cumulative-months=7;msg-param-months=0;msg-param-should-share-streak=1;msg-param-streak-months=8;msg-param-sub-plan-name=Channel\\sSubscription\\s(loeya);msg-param-sub-plan=1000;room-id=166279350;subscriber=1;system-msg=Ovojaytee\\ssubscribed\\sat\\sTier\\s1.\\sThey've\\ssubscribed\\sfor\\s8\\smonths,\\scurrently\\son\\sa\\s8\\smonth\\sstreak!;tmi-sent-ts=1558352544376;user-id=160605648;user-type= :tmi.twitch.tv USERNOTICE #loeya :Wow 8 months loeyaH our baby is almost here loeyaHM can we name him Zlatan ? Thanks Queen for always starting off my day on a good note with your wonderful content loeya1\r\n")
	scanner := NewScanner(reader)

	msg, err := scanner.Scan()

	assert.NoError(t, err)

	assert.Equal(t, &Message{
		Tags: map[string]string{
			"badge-info":                    "subscriber/8",
			"badges":                        "subscriber/6,bits/75000",
			"color":                         "#1E90FF",
			"display-name":                  "Ovojaytee",
			"emotes":                        "1837404:44-50/915234:164-169/1093027:13-18",
			"flags":                         "",
			"id":                            "aa52e1d2-6ff5-42ba-b205-9d4a15f9dbf8",
			"login":                         "ovojaytee",
			"mod":                           "0",
			"msg-id":                        "resub",
			"msg-param-cumulative-months":   "7",
			"msg-param-months":              "0",
			"msg-param-should-share-streak": "1",
			"msg-param-streak-months":       "8",
			"msg-param-sub-plan":            "1000",
			"msg-param-sub-plan-name":       "Channel Subscription (loeya)",
			"room-id":                       "166279350",
			"subscriber":                    "1",
			"system-msg":                    "Ovojaytee subscribed at Tier 1. They've subscribed for 8 months, currently on a 8 month streak!",
			"tmi-sent-ts":                   "1558352544376",
			"user-id":                       "160605648",
			"user-type":                     "",
		},
		Prefix:  "tmi.twitch.tv",
		Command: "USERNOTICE",
		Params: []string{
			"#loeya",
			"Wow 8 months loeyaH our baby is almost here loeyaHM can we name him Zlatan ? Thanks Queen for always starting off my day on a good note with your wonderful content loeya1",
		},
	}, msg)
}
