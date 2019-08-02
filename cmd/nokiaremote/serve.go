package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	gocli "gopkg.in/src-d/go-cli.v0"
)

func init() {
	app.AddCommand(&ServeCommand{})
}

type ServeCommand struct {
	gocli.PlainCommand `name:"serve" short-description:"start server" long-description:"Starts an HTTP server for the app to connect to"`
	Port               string `long:"port" env:"NOKIAREMOTE_PORT" description:"Port to listen on" default:"80"`

	broadcast  chan string              // Inbound actions from the app.
	clients    map[*websocket.Conn]bool // Registered clients.
	register   chan *websocket.Conn     // Register requests from the clients.
	unregister chan *websocket.Conn     // Unregister requests from clients.

	upgrader websocket.Upgrader
}

func (s *ServeCommand) ExecuteContext(ctx context.Context, args []string) error {
	s.broadcast = make(chan string)
	s.clients = map[*websocket.Conn]bool{}
	s.register = make(chan *websocket.Conn)
	s.unregister = make(chan *websocket.Conn)
	s.upgrader = websocket.Upgrader{}

	e := echo.New()
	e.GET("/action", s.serveAction)
	e.Any("/socket", s.serveSocket)
	e.GET("/wap.wml", s.serveWML)

	go s.runBroadcaster()

	go func() {
		<-ctx.Done()
		e.Shutdown(ctx)
	}()

	return e.Start(fmt.Sprintf(":%s", s.Port))
}

func (s *ServeCommand) serveAction(c echo.Context) error {
	action := c.QueryParams().Get("action")
	if action == "" {
		return c.String(http.StatusBadRequest, "no action defined")
	}
	s.broadcast <- action

	if c.QueryParams().Get("wap") == "1" {
		return c.Redirect(http.StatusTemporaryRedirect, "wap.wml")
	}

	return c.String(http.StatusOK, "OK")
}

func (s *ServeCommand) serveSocket(c echo.Context) error {
	sock, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return err
	}
	defer func() { s.unregister <- sock }()
	s.register <- sock
	for {
		_, message, err := sock.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		log.Printf("recv: %s", message)
	}

	return nil
}

func (s *ServeCommand) serveWML(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")
	return c.File("./wap.wml")
}

func (s *ServeCommand) runBroadcaster() {
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				client.Close()
			}
		case message := <-s.broadcast:
			for client := range s.clients {
				client.WriteMessage(websocket.TextMessage, []byte(message))
			}
		}
	}
}
