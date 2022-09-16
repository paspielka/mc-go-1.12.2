package gomcbot

import (
	// "fmt"
	//"github.com/mattn/go-colorable"//On Windows need
	"testing"
)

/*
	"translation.test.none":                                      "Hello, world!",
	"translation.test.complex":                                   "Prefix, %s %[2]s again %s   and %[1]s lastly %s   and also %[1]s again!"
																  "Prefix, str1str2 again str3 and str1  lastly str2 and also str1  again!"
																  "Prefix, str1str2 again str2 and str1  lastly str3 and also str1  again!"
	"translation.test.escape":                                    "%%s %%%s %%%%s %%%%%s",
	"translation.test.invalid":                                   "hi %",
	"translation.test.invalid2":                                  "hi %  s",
	"translation.test.args":                                      "%s %s",
	"translation.test.world":
*/
var jsons = []string{
	`{"extra":[{"color":"green","text":"故我依然"},{"color":"white","text":"™ "},{"color":"gray","text":"Kun_QwQ"},{"color":"white","text":": 为什么想要用炼药锅灭火时总是跳不进去"}],"text":""}`,

	`{"translate":"chat.type.text","with":[{"insertion":"Xi_Xi_Mi","clickEvent":{"action":"suggest_command","value":"/tell Xi_Xi_Mi "},"hoverEvent":{"action":"show_entity","value":{"text":"{name:\"{\\\"text\\\":\\\"Xi_Xi_Mi\\\"}\",id:\"c1445a67-7551-4d7e-813d-65ef170ae51f\",type:\"minecraft:player\"}"}},"text":"Xi_Xi_Mi"},"好像是这个id。。"]}`,
	`{"translate":"translation.test.none"}`,
	//`{"translate":"translation.test.complex","with":["str1","str2","str3"]}`,
	`{"translate":"translation.test.escape","with":["str1","str2"]}`,
	`{"translate":"translation.test.args","with":["str1","str2"]}`,
	`{"translate":"translation.test.world"}`,
}

var texts = []string{
	"\033[92m故我依然\033[0m\033[97m™ \033[0m\033[37mKun_QwQ\033[0m\033[97m: 为什么想要用炼药锅灭火时总是跳不进去\033[0m",

	"<Xi_Xi_Mi> 好像是这个id。。",
	"Hello, world!",
	//"Prefix, str1str2 again str2 and str1 lastly str3 and also str1 again!",
	"%s %str1 %%s %%str2",
	"str1 str2",
	"world",
}

func TestChatMsgFormatString(t *testing.T) {
	for i, v := range jsons {
		cm, err := newChatMsg([]byte(v))
		if err != nil {
			t.Error(err)
		}
		if cm.String() != texts[i] {
			t.Errorf("%v Should be %s", cm, texts[i])
		}
	}
}
