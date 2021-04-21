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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// InitTunnel -
// ATTEMPTS TO CREATE NGROK TUNNEL
func (c *Client) InitTunnel(t *Tunnel) (err error) {
	for attempt := uint8(0); attempt <= Settings.MaxRetries; attempt++ {
		err = func() error {
			if Settings.ShouldLog {
				Logger.Println("Attempting to initialize tunnel...")
				Logger.Printf(">>>\tName: %s | Addr: %s\n", t.Name, t.LocalAddress)
			}
			time.Sleep(1 * time.Second)

			jsonData := t.getJSON()

			if protocols[t.Proto] == "http" {
				// IF HTTP APPEND BIND_TLS
				jsonData["bind_tls"] = true
			}

			url := fmt.Sprintf(Settings.TunnelAPIAddr, c.NGROKLocalAddr)

			jsonValue, err := json.Marshal(jsonData)
			if err != nil {
				return err
			}

			// INIT RECORD
			record, err := attemptInitTunnel(url, jsonValue)

			if err != nil {
				if Settings.ShouldLog {
					Logger.Printf("failed to init tunnel: %s | addr: %s\n", t.Name, t.LocalAddress)
				}
				return err
			}

			t.RemoteAddress = record.PublicURL
			t.IsCreated = true

			if Settings.ShouldLog {
				Logger.Println("Tunnel Created...")
				Logger.Printf(">>> Name: %s | Public Addr: %s\n", t.Name, t.RemoteAddress)
			}
			return nil
		}()
		if c.LogAPI && err != nil {
			if Settings.ShouldLog {
				Logger.Println(err)
			}
		}
		if err == nil {
			break
		}
	}
	return
}

// attemptInitTunnel
// NGROK REQUEST TO INIT TUNNEL
func attemptInitTunnel(url string, jsonBytes []byte) (*ngrokTunnelRecord, error) {
	record := &ngrokTunnelRecord{}
	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		res, _ := ioutil.ReadAll(res.Body)
		return nil, errors.New("error api: " + string(res))
	}

	if err := json.NewDecoder(res.Body).Decode(&record); err != nil {
		return nil, err
	}
	return record, nil
}

// CloseTunnel -
// CLOSE NGROK TUNNEL
func (c *Client) CloseTunnel(t *Tunnel) (err error) {
	for attempt := uint8(0); attempt <= Settings.MaxRetries; attempt++ {
		err = func() error {
			if Settings.ShouldLog {
				Logger.Println("Closing ngrok tunnel...")
				Logger.Printf(">>> Addr: %s | Local Server Addr: %s", t.RemoteAddress, t.LocalAddress)
			}

			url := fmt.Sprintf("%s/%s", fmt.Sprintf(Settings.TunnelAPIAddr, c.NGROKLocalAddr), t.Name)
			err := attemptCloseTunnel(url)
			if err != nil {
				if Settings.ShouldLog {
					Logger.Printf("attemptCloseTunnel err: %s\n", err)
				}
				return err
			}
			t.RemoteAddress = ""
			t.IsCreated = false
			if Settings.ShouldLog {
				Logger.Println("Successfully closed tunnel...")
				Logger.Printf(">>> Closed tunnel name: %s\n", t.Name)
			}
			return nil
		}()
		if c.LogAPI && err != nil {
			if Settings.ShouldLog {
				Logger.Println(err)
			}
		}
		if err == nil {
			break
		}
	}
	return
}

// attemptCloseTunnel -
// ATTEMPT TO CLOSE TUNNEL W/ PROVIDED URL
func attemptCloseTunnel(url string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		if Settings.ShouldLog {
			Logger.Println(err)
		}
		return err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		if Settings.ShouldLog {
			Logger.Println(err)
		}
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		res, _ := ioutil.ReadAll(res.Body)
		return errors.New("error api: " + string(res))
	}
	return nil
}
