package gongrok

/*
	Copyright 2021 revzim.

	Permission is hereby granted, free of charge, to any person obtaining a
	copy of this software and associated documentation files (the "Software"),
	to deal in the Software without restriction, including without limitation
	the rights to use, copy, modify, merge, publish, distribute, sublicense,
	and/or sell copies of the Software, and to permit persons to whom the Software
	is furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included
	in all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
	EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
	OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
	IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
	DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
	ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
	DEALINGS IN THE SOFTWARE.

*/
import (
	"log"
	"os/exec"
)

type (
	// Map -
	// TYPEALIAS FOR map[string]interface{}
	Map map[string]interface{}

	// Protocol -
	// ALIAS FOR SUPPORTED PROTOCOLS
	Protocol int

	// ngrokTunnelRecord
	// DATA ABOUT CURRENT NGROK TUNNEL
	ngrokTunnelRecord struct {
		Name      string  `json:"Name"`       // NGROK TUNNEL IDENTIFYING NAME
		URI       string  `json:"uri"`        // URI
		PublicURL string  `json:"public_url"` // NGROK PUBLIC URL
		Proto     string  `json:"Proto"`      // PROTOCOL 0, 1, 2 | HTTP, TCP, TLS
		Config    Config  `json:"config"`     // TUNNEL CONFIG
		Metricts  Metrics `json:"metrics"`    // TUNNEL METRICS
	}

	// Config -
	// ResponseInitTunnel CONFIGURATION
	Config struct {
		Addr    string `json:"addr"`    // ADDRESS
		Inspect bool   `json:"Inspect"` // SHOULD INSPECT TRANSACTIONAL DATA OF NGROK TUNNEL
	}

	// Metrics -
	// ResponseInitTunnel METRICS
	Metrics struct {
		Conns info `json:"conns"` // NGROK CONNECTIONS DATA
		HTTP  info `json:"http"`  // NGROK HTTP DATA
	}

	// info -
	// NGROK INFO
	info struct {
		Count  int `json:"count"`
		Gauge  int `json:"gauge"`
		Rate1  int `json:"rate1"`
		Rate5  int `json:"rate5"`
		Rate15 int `json:"rate15"`
		P50    int `json:"p50"`
		P90    int `json:"p90"`
		P95    int `json:"p95"`
		P99    int `json:"p99"`
	}

	// Tunnel -
	// INIT/CLOSE TUNNEL
	// AUTO-CONNECT TO NGROK IF SERVER IS UP
	Tunnel struct {
		Proto         Protocol `json:"proto"`      // PROTOCOL 0, 1, 2 | HTTP, TCP, TLS
		Name          string   `json:"name"`       // TUNNEL NAME IDENTIFIER
		LocalAddress  string   `json:"localaddr"`  // HOST | HOST:PORT
		Auth          string   `json:"auth"`       // AUTH FOR TUNNEL, IF ANY
		Inspect       bool     `json:"inspect"`    // INSPECT TRANSACTIONAL DATA OF NGROK TUNNEL
		RemoteAddress string   `json:"remoteaddr"` // NGROK PUBLIC ADDRESS
		IsCreated     bool     `json:"iscreated"`  // IF TUNNEL CREATED
	}
	// Options -
	// OPTIONS FOR COMMAND TO START NGROK
	Options struct {
		SubDomain string `json:"subdomain"` // SUBDOMAIN *PREMIUM*
		AuthToken string `json:"authtoken"` // AUTH TOKEN
		Region    string `json:"region"`    // TUNNEL REGION
		CFGPath   string `json:"cfgpath"`   // NGROK CFG PATH
		NGROKPath string `json:"binpath"`   // NGORK BIN PATH
		LogNGROK  bool   `json:"logbin"`    // SHOULD LOG NGROK BIN OR NOT
	}

	// Client -
	// NGROK CLIENT USED FOR MONITORING TUNNEL CREATION/DELETION & MORE
	Client struct {
		ID             string    `json:"id"`                   // IDENTIFIER FOR CLIENT
		Options        *Options  `json:"options"`              // CMD OPTIONS
		Tunnels        []*Tunnel `json:"tunnels"`              // ALL CLIENT TUNNELS
		NGROKLocalAddr string    `json:"ngroklocaladdr"`       // CLIENT LOCAL SERVER FOR NGROK METRICS/API
		LogAPI         bool      `json:"logapi"`               // SHOULD LOG API RESPONSE
		cmds           []string  `json:"commands,omitempty"`   // CMDS USED TO RUN NGROKBIN
		runningCMDS    *exec.Cmd `json:"runningcmd,omitempty"` // RUNNING CMDS
	}

	// settings -
	// GONGROK SETTINGS
	settings struct {
		Path          string `json:"path"`          // PATH TO NGROK BINARY FILE (https://ngrok.com/download)
		DefaultPath   string `json:"default_path"`  // PATH TO NGROK BINARY FILE (https://ngrok.com/download)
		LogDir        string `json:"logdir"`        // DIRECTORY WHERE USER WANTS LOGS TO POPULATE
		LogAPI        bool   `json:"logapi"`        // SHOULD LOG API OR NOT
		ShouldLog     bool   `json:"shouldlog"`     // SHOULD LOG ANYTHING
		MaxRetries    uint8  `json:"maxretries"`    // HOW MANY RETRIES OF TUNNEL CREATION/DELETION
		TunnelAPIAddr string `json:"tunnelAPIaddr"` // ADDRESS OF NGROK TUNNEL API
	}
)

const (
	ngReady          = `starting web service.*addr=(\d+\.\d+\.\d+\.\d+:\d+)`   // IS NGROK READY
	ngInUse          = `address already in use`                                // IS PORT IN USE
	ngSessionLimited = `is limited to (\d+) simultaneous ngrok client session` // CHECK NGROK LIMIT
	webURI           = `\d+\.\d+\.\d+\.\d+:\d+`                                // FIND NGROK CLIENT SERVER
)

// SUPPORTED PROTOCOLS
const (
	HTTP Protocol = iota
	TCP
	TLS
)

var (
	protocols = map[Protocol]string{
		HTTP: "http",
		TCP:  "tcp",
		TLS:  "tls",
	}
	// Settings -
	// NGROK DEFAULT SETTINGS
	Settings = settings{
		Path:          "./ngrok_bin/ngrok",
		DefaultPath:   "./ngrok_bin/ngrok",
		LogDir:        "./logs",
		LogAPI:        false,
		ShouldLog:     false,
		MaxRetries:    50,
		TunnelAPIAddr: "http://%s/api/tunnels",
	}
	// Logger -
	// LOGGER
	Logger *log.Logger
)

func (t *Tunnel) getJSON() Map {
	return Map{
		"addr":    t.LocalAddress,
		"proto":   protocols[t.Proto],
		"name":    t.Name,
		"inspect": t.Inspect,
		"auth":    t.Auth,
	}
}
