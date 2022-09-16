package test

import (
	"fmt"
	bot "github.com/edouard127/mc-go-1.12.2"
	"github.com/edouard127/mc-go-1.12.2/util"
	"testing"
	"time"
)

func TestCalibrate(t *testing.T) {
	p := bot.Auth{
		Name: "Name",
		UUID: "UUID",
		AsTk: "",
	}
	g, err := p.JoinServer("localhost", 25565)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Login success")
	events := g.GetEvents()
	go g.HandleGame()

	for e := range events {
		switch e.(type) {
		case bot.PlayerSpawnEvent:
			fmt.Println("Player Spawn")
			go func() {
				for {
					gomcbot.CalibratePos(g)
					time.Sleep(2 * time.Second)
				}
			}()

		default:
			// fmt.Println(e)
		}
	}
}
