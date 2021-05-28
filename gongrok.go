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
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// InitLoggerWriter -
// INIT LOGGER FOR WRITING TO FILE/STDOUT
func InitLoggerWriter(fileName string) {
	err := ngrokBinExists()
	if err != nil {
		log.Fatalf("no ngrok binary in path <Settings.Path>: %s", Settings.Path)
	}
	t := time.Now()
	formatted := t.Format("2006_02_01__15_04_05")
	fileName = fmt.Sprintf("%s_%s.log", fileName, formatted)
	if _, err := os.Stat(Settings.LogDir); os.IsNotExist(err) {
		os.Mkdir(Settings.LogDir, 0700)
	}
	f, err := os.OpenFile(fmt.Sprintf("./logs/%s", fileName), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)
	Logger = log.New(wrt, fmt.Sprintf("gongrok | %s | > ", t.Format("2006_02_01__15_04_05")), 0)
	Logger.Println("Logger Init...")
}

// ngrokBinExists
func ngrokBinExists() error {
	if runtime.GOOS == "windows" {
		Settings.Path = fmt.Sprintf("%s.exe", Settings.DefaultPath)
	}
	f, err := os.OpenFile(Settings.Path, os.O_RDWR, 0644)
	if errors.Is(err, os.ErrNotExist) {
		return err
	}
	defer f.Close()

	return nil
}

// NewClient -
// INITS & RETURNS NEW CLIENT
func NewClient(opt Options) (*Client, error) {
	err := ngrokBinExists()
	if err != nil {
		log.Fatalf("no ngrok binary in path <Settings.Path>: %s", Settings.Path)
	}
	if Settings.ShouldLog {
		Logger.Println("New client")
	}

	if opt.NGROKPath == "" {
		opt.NGROKPath = Settings.Path
	}

	if opt.Region == "" {
		opt.Region = "us"
	}

	if opt.AuthToken != "" {
		err := opt.AuthTokenCommand()
		if err != nil {
			return nil, err
		}
	}

	c := &Client{ID: uuid.New().String(), Options: &opt, LogAPI: Settings.LogAPI}
	return c, nil
}

// AuthTokenCommand -
// CMD USED IF AUTHENTICATION PROVIDED
func (o *Options) AuthTokenCommand() error {
	if o.AuthToken == "" {
		return errors.New("token missing")
	}

	if o.NGROKPath == "" {
		return errors.New("binary path file is missing")
	}

	commands := make([]string, 0)
	commands = append(commands, []string{"authtoken", o.AuthToken}...)

	if o.CFGPath != "" {
		commands = append(commands, "--config"+o.CFGPath)
	}

	cmd := exec.Command(o.NGROKPath, commands...)
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	if err := cmd.Start(); err != nil {
		return err
	}

	if errBuffer.String() != "" {
		return errors.New(errBuffer.String())
	}
	if Settings.ShouldLog {
		Logger.Println(outBuffer.String())
	}
	return nil
}

// StartNGROK -
// PARSE NGROK RESPONSE & CLOSE WAITGROUP ONCE RECVD
// PAYLOAD THAT NGROK SERVER CLIENT READY
func (c *Client) StartNGROK(wg *sync.WaitGroup) error {
	if Settings.ShouldLog {
		Logger.Println("Start server")
	}
	commands := c.Options.generateCommands()
	cmd := exec.Command(c.Options.NGROKPath, commands...)
	c.runningCMDS = cmd
	out, err := cmd.StdoutPipe()
	if err != nil {
		if Settings.ShouldLog {
			// Logger.Fatal(err)
			Logger.Printf("Pipe cmd err: %s", err.Error())
			return err
		}
		return err
	}

	if err := cmd.Start(); err != nil {
		if Settings.ShouldLog {
			// Logger.Fatal(err)
			Logger.Printf("Start cmd err: %s", err.Error())
			return err
		}
	}

	// HANDLES SIGNAL INPUT
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan, syscall.SIGHUP,
		syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT)
	go c.handleSignalInput(signalChan)

	// ATTEMPT TO INITIAILIZE NGROK CLIENT SERVER
	err = handleInitNGROK(wg, c, out)

	if err != nil {
		Logger.Println("handleInitNGROK error:", err)
		return err
	}

	return nil
}

// handleInitNGROK -
// INIT REGEXES FOR NGROK CLIENT SERVER RESPONSE
// IF REGEX PASSES, ATTEMPT TO PARSE RESPONSE
func handleInitNGROK(wg *sync.WaitGroup, c *Client, out io.ReadCloser) error {
	isNGReady, err := regexp.Compile(ngReady)
	if err != nil {
		if Settings.ShouldLog {
			Logger.Printf("check ngrok read err: %s", err.Error())
		}
		return err
	}

	isNGInUse, err := regexp.Compile(ngInUse)
	if err != nil {
		if Settings.ShouldLog {
			Logger.Printf("check ngrok in use err: %s", err.Error())
		}
		return err
	}

	isNGSessionLimit, err := regexp.Compile(ngSessionLimited)
	if err != nil {
		if Settings.ShouldLog {
			Logger.Printf("check ngrok session limit err: %s", err.Error())
		}
		return err
	}

	isLocalNGROKURI, err := regexp.Compile(webURI)
	if err != nil {
		if Settings.ShouldLog {
			Logger.Printf("check local ngrok uri err: %s\n", err.Error())
		}
		return err
	}
	err = parseNGROK(wg, isNGReady, isLocalNGROKURI, isNGInUse, isNGSessionLimit, c, out)
	if err != nil {
		if Settings.ShouldLog {
			Logger.Printf("parse ngrok error: %s\n", err.Error())
		}
		return err
	}
	return nil
}

// parseNGROK -
// REGEX MATCHES & PARSES NGROK RESPONSE
// ON SUCCESS, YIELDS NGROK CLIENT SERVER PUBLIC ADDR
func parseNGROK(wg *sync.WaitGroup, isNGReady, isLocalNGURI, isNGInUse, isNGSessionLimit *regexp.Regexp, c *Client, out io.ReadCloser) error {

	chunk := make([]byte, 256)
	for {
		n, err := out.Read(chunk)
		if err != nil {
			if Settings.ShouldLog {
				Logger.Printf("check ngrok read err: %s", err.Error())
			}
			return err
		}

		if n < 1 {
			continue
		}

		if c.Options.LogNGROK && Settings.ShouldLog {
			Logger.Print("NGROK LOG: ", string(chunk[:n]))
		}
		// REGEX OUTPUT SEARCHES LOCAL IP & PORT FOR NGROK WEB UI
		if isNGReady.Match(chunk[:n]) {
			host := isLocalNGURI.FindStringSubmatch(string(chunk[:n]))
			if len(host) >= 1 {
				if Settings.ShouldLog {
					Logger.Println("server client ready")
				}
				c.NGROKLocalAddr = host[0]
				wg.Done()
			}
		}
		if isNGInUse.Match(chunk[:n]) {
			if Settings.ShouldLog {
				Logger.Printf("ngrok addr already in use: %s", err.Error())

			}
			return err
		}
		if isNGSessionLimit.Match(chunk[:n]) {
			if Settings.ShouldLog {
				Logger.Printf("ngrok session limit reached")
			}
			return errors.New("ngrok session limit reached")
		}
	}
}

// generateCommands -
// RETURNS COMMANDS TO START NGROK BIN
func (o *Options) generateCommands() []string {
	cmds := []string{"start", "--none", "--log=stdout", fmt.Sprintf("--region=%s", o.Region)}

	if o.CFGPath != "" {
		cmds = append(cmds, fmt.Sprintf("--cfg=%s", o.CFGPath))
	}
	if o.SubDomain != "" {
		cmds = append(cmds, fmt.Sprintf("--subdomain=%s", o.SubDomain))
	}

	return cmds
}

// handleSignalInput -
// HANDLES SIGNAL INPUT
func (c *Client) handleSignalInput(signalChan chan os.Signal) {
	for {
		s := <-signalChan
		switch s {
		default:
			if Settings.ShouldLog {
				Logger.Println(s)
			}
			c.Signal(s)
			os.Exit(1)
		}
	}
}

// AddTunnel -
// INIT NEW TUNNEL
func (c *Client) AddTunnel(t *Tunnel) {
	if Settings.ShouldLog {
		Logger.Println("Add tunnel")
	}
	c.Tunnels = append(c.Tunnels, t)
}

// ConnectAll -
// CONNECT ALL TUNNELS FOR CLIENT
func (c *Client) ConnectAll() error {
	wg := &sync.WaitGroup{}
	// NGROK TUNNELS API REQUESTS POST TO API/TUNNELS
	if Settings.ShouldLog {
		Logger.Println("Connecting")
	}
	if len(c.Tunnels) < 1 {
		return errors.New("client currently has 0 tunnels")
	}

	for _, tunnel := range c.Tunnels {
		if !tunnel.IsCreated {
			wg.Add(1)
			go func(tunnel *Tunnel) {
				c.InitTunnel(tunnel)
				wg.Done()
			}(tunnel)

		}
	}

	wg.Wait()
	return nil
}

// DisconnectTunnel -
// DISCONNECT SPECIFIED TUNNEL NAME FROM CLIENT
func (c *Client) DisconnectTunnel(name string) error {
	wg := &sync.WaitGroup{}
	if Settings.ShouldLog {
		Logger.Println("Disconnecting")
	}
	if len(c.Tunnels) < 1 {
		return errors.New("client currently has 0 tunnels")
	}

	for _, clientTunnel := range c.Tunnels {
		if clientTunnel.IsCreated && clientTunnel.Name == name {
			wg.Add(1)
			go func() {
				c.CloseTunnel(clientTunnel)
				wg.Done()
			}()
		}
	}
	return nil
}

// DisconnectAll -
// DISCONNECT ALL CLIENT TUNNELS
func (c *Client) DisconnectAll() error {
	wg := &sync.WaitGroup{}
	//	api request delete to /api/tunnels/:Name
	if Settings.ShouldLog {
		Logger.Println("Disconnecting...")
	}
	if len(c.Tunnels) < 1 {
		return errors.New("client currently has 0 tunnels")
	}

	for _, t := range c.Tunnels {
		if t.IsCreated {
			wg.Add(1)
			go func() {
				c.CloseTunnel(t)
				wg.Done()
			}()
		}
	}

	wg.Wait()
	return nil
}

// Close -
// CLOSE & KILL NGROK CMD
func (c *Client) Close() error {
	return c.runningCMDS.Process.Kill()
}

// Signal -
// HANDLE SIGINPUT
func (c *Client) Signal(signal os.Signal) error {
	return c.runningCMDS.Process.Signal(signal)
}
