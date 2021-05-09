package parser

import (
	"bytes"
	"testing"
)

var res *Message

func BenchmarkScanner_Scan(b *testing.B) {
	b.StopTimer()
	var r *Message
	var buf bytes.Buffer
	scanner := NewScanner(&buf)
	var err error
	b.Run("Long Scan", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			b.StopTimer()
			buf.WriteString("@badge-info=subscriber/8;badges=subscriber/6,bits/75000;color=#1E90FF;display-name=Ovojaytee;emotes=1837404:44-50/915234:164-169/1093027:13-18;flags=;id=aa52e1d2-6ff5-42ba-b205-9d4a15f9dbf8;login=ovojaytee;mod=0;msg-id=resub;msg-param-cumulative-months=7;msg-param-months=0;msg-param-should-share-streak=1;msg-param-streak-months=8;msg-param-sub-plan-name=Channel\\sSubscription\\s(loeya);msg-param-sub-plan=1000;room-id=166279350;subscriber=1;system-msg=Ovojaytee\\ssubscribed\\sat\\sTier\\s1.\\sThey've\\ssubscribed\\sfor\\s8\\smonths,\\scurrently\\son\\sa\\s8\\smonth\\sstreak!;tmi-sent-ts=1558352544376;user-id=160605648;user-type= :tmi.twitch.tv USERNOTICE #loeya :Wow 8 months loeyaH our baby is almost here loeyaHM can we name him Zlatan ? Thanks Queen for always starting off my day on a good note with your wonderful content loeya1\r\n")
			b.StartTimer()
			r, err = scanner.Scan()
			if r == nil {
				b.Errorf("Received nil message from scan: %v", err)
			}
		}
		res = r
	})
	b.Run("Short Scan", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			b.StopTimer()
			buf.WriteString(":tmi.twitch.tv CAP * ACK :twitch.tv/tags twitch.tv/commands twitch.tv/membership\r\n")
			b.StartTimer()
			r, err = scanner.Scan()
			if r == nil {
				b.Errorf("Received nil message from scan: %v", err)
			}
		}
		res = r
	})
}
