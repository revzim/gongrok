package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo"

	"github.com/revzim/gongrok"
)

type (
	str string
)

var (
	clients map[string]*gongrok.Client
)

func handleDisconnectTunnel(c echo.Context) error {
	clientID := c.FormValue("clientid")
	tunnelName := c.FormValue("tunnelname")
	if gongrok.Settings.ShouldLog {
		gongrok.Logger.Printf("handleDisconnectTunnel >>> Attemping to disconnect client | tunnel: %s | %s\n", clientID, tunnelName)
	}

	var client *gongrok.Client
	var ok bool
	if client, ok = clients[clientID]; !ok {
		return c.JSON(http.StatusOK, echo.Map{
			"error": fmt.Errorf("No client with id %s exists", clientID),
			"code":  200,
		})
	}
	// CLIENT EXISTS
	if gongrok.Settings.ShouldLog {
		gongrok.Logger.Printf("Client %s exists\nRemoving tunnel %s...", clientID, tunnelName)
	}
	for i := range client.Tunnels {
		if client.Tunnels[i].Name == tunnelName {
			removedTunnel := client.Tunnels[i]
			err := client.DisconnectTunnel(tunnelName)
			if err != nil {
				return c.JSON(http.StatusOK, echo.Map{
					"error": err,
					"code":  200,
				})
			}
			return c.JSON(http.StatusOK, echo.Map{
				"code":    200,
				"status":  "OK",
				"removed": removedTunnel,
			})
		}
	}
	return c.JSON(http.StatusFound, echo.Map{
		"error":  fmt.Errorf("tunnel %s does not belong to %s", tunnelName, clientID),
		"code":   200,
		"status": "fail",
	})
}

func handleDisconnectClient(c echo.Context) error {
	clientID := c.FormValue("clientid")
	if gongrok.Settings.ShouldLog {
		gongrok.Logger.Printf("handleDisconnectClient >>> Attemping to disconnect client: %s\n", clientID)
	}
	if _, ok := clients[clientID]; ok {
		if gongrok.Settings.ShouldLog {
			gongrok.Logger.Printf("Client %s exists\nRemoving...", clientID)
		}
		err := clients[clientID].DisconnectAll()
		if err != nil {
			return c.JSON(http.StatusOK, echo.Map{
				"error": err,
				"code":  200,
			})
		}
	}

	err := clients[clientID].Close()
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{
			"error":  err,
			"code":   200,
			"status": "FAIL",
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"code":   200,
		"status": "OK",
	})
}

func handleNewClient(c echo.Context) error {
	protocolstr := str(c.FormValue("protocol"))
	protocol, err := protocolstr.toInt()
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{
			"error": err,
			"code":  200,
		})
	}
	if checkProtocol(protocol) {
		return c.JSON(http.StatusOK, echo.Map{
			"error": fmt.Sprintf("client sent bad protocol: %d", protocol),
			"code":  200,
		})
	}
	portstr := str(c.FormValue("port"))
	var port int
	port, err = portstr.toInt()
	if err != nil || port == -1 {
		return c.JSON(http.StatusOK, echo.Map{
			"error": err,
			"code":  200,
		})
	}
	clientname := c.FormValue("tunnelname")
	hostname := c.FormValue("host")
	if gongrok.Settings.ShouldLog {
		gongrok.Logger.Println("port:", port, "protocol:", protocol, "host:", hostname, "tunnelname:", clientname)
	}
	tunnel := &gongrok.Tunnel{
		Proto:        gongrok.Protocol(protocol),
		Name:         clientname,
		LocalAddress: fmt.Sprintf("%s:%d", hostname, port),
		Auth:         "",
	}

	client, err := gongrok.NewClient(gongrok.Options{
		LogNGROK: true,
	})
	if err != nil {
		if gongrok.Settings.ShouldLog {
			gongrok.Logger.Fatal(err)
		}
		return c.JSON(http.StatusOK, echo.Map{
			"error": err,
			"code":  200,
		})
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		err := client.StartNGROK(wg)
		if err != nil {
			if gongrok.Settings.ShouldLog {
				gongrok.Logger.Println("server err:", err, " (*disregard if EOF*)")
			}
		}
	}(wg)
	wg.Wait()

	client.AddTunnel(tunnel)

	err = client.ConnectAll()

	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{
			"error": err,
			"code":  200,
		})
	}
	if gongrok.Settings.ShouldLog {
		gongrok.Logger.Printf("CLIENT CONNECTED: %+v\n", client)
	}
	clients[client.ID] = client

	return c.JSON(http.StatusOK, echo.Map{
		"tunnel": tunnel,
		"client": client,
	})

}

func checkProtocol(protocol int) bool {
	switch protocol {
	case 0:
		fallthrough
	case 1:
		fallthrough
	case 2:
		return false
	default:
		break
	}
	return true
}

func (s str) toInt() (int, error) {
	var err error
	port, err := strconv.Atoi(string(s))
	if err != nil {
		return -1, nil
	}
	return port, err
}

func handleClientHome(c echo.Context) error {
	c.Response().Header().Add(echo.HeaderCookie, "POOP")
	return c.File("./public/index.html")
}

func main() {
	// IF WANT TO WRITE TO FILE
	clients = make(map[string]*gongrok.Client)
	gongrok.Settings.LogAPI = true
	gongrok.Settings.ShouldLog = true
	gongrok.InitLoggerWriter("test")
	e := echo.New()

	e.Logger.SetOutput(gongrok.Logger.Writer())
	e.Static("/public", "public")
	e.GET("/client", handleClientHome)

	e.POST("/client/new", handleNewClient)
	e.POST("/client/disconnect", handleDisconnectClient)
	// e.POST("/client/tunnel/disconnect", handleDisconnectTunnel)

	// SILLY DELAY TO PRINT EASY CLIENT ADDR
	time.AfterFunc(1*time.Second, func() {
		gongrok.Logger.Println("\nGONGROK CLIENT: http://localhost:8080/client")
	})

	gongrok.Logger.Fatal(e.Start(":8080"))
}
