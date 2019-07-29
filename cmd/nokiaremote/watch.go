package main

import (
	"context"
	"log"
	"net/url"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"github.com/micmonay/keybd_event"
	gocli "gopkg.in/src-d/go-cli.v0"
)

func init() {
	app.AddCommand(&WatchCommand{})
}

type WatchCommand struct {
	gocli.PlainCommand `name:"watch" short-description:"start server" long-description:"Starts an HTTP server for the app to connect to"`
	Host               string `long:"host" env:"NOKIAREMOTE_HOST" description:"Hostname to connect to" default:"nokiaremote.maartje.dev"`
	Path               string `long:"path" env:"NOKIAREMOTE_PATH" description:"Path on the host to connect to" default:"/socket"`

	broadcast  chan string              // Inbound actions from the app.
	clients    map[*websocket.Conn]bool // Registered clients.
	register   chan *websocket.Conn     // Register requests from the clients.
	unregister chan *websocket.Conn     // Unregister requests from clients.

	upgrader websocket.Upgrader
}

func (s *WatchCommand) ExecuteContext(ctx context.Context, args []string) error {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return err
	}

	// For linux, it is very important wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	u := url.URL{Scheme: "ws", Host: s.Host, Path: s.Path}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	for {
		select {
		case <-ctx.Done():
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
			}
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
			}
			log.Printf("recv: %s", message)
			launchKey := false
			if string(message) == "next" {
				kb.SetKeys(keybd_event.VK_RIGHT)
				launchKey = true
			}
			if string(message) == "previous" {
				kb.SetKeys(keybd_event.VK_LEFT)
				launchKey = true
			}
			if string(message) == "start" {
				kb.SetKeys(keybd_event.VK_F5)
				kb.HasCTRL(true)
				launchKey = true
			}
			if string(message) == "stop" {
				kb.SetKeys(keybd_event.VK_ESC)
				launchKey = true
			}

			if launchKey {
				err = kb.Launching()
				if err != nil {
					log.Println(err)
				}
				//reseting
				kb.HasCTRL(false)
			}
		}
	}
}
