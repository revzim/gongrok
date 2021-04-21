# gongrok
[![Go Report Card](https://goreportcard.com/badge/github.com/revzim/gongrok)](https://goreportcard.com/report/github.com/revzim/gongrok)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/revzim/gongrok)](https://pkg.go.dev/github.com/revzim/gongrok)
[![GoDoc](https://godoc.org/github.com/revzim/gongrok?status.svg)](https://godoc.org/github.com/revzim/gongrok)

## golang [ngrok](https://ngrok.com/) wrapper
  * `Secure introspectable tunnels to local host`

## Install

 *  [ngrok](https://ngrok.com/download) binary is required for use
    *  download the binary
    *  place binary within your working directory
    *  ```default: ./ngrok_bin/```
    
 * ``` go get github.com/revzim/gongrok ```

## Example

![Web App Image](https://i.imgur.com/XEerXOm.png)
* [WEB APP EXAMPLE](https://github.com/revzim/gongrok/example/webapp)
  * place ngrok binary download in ngrok path ```./ngrok_bin/```
  * build or run
    * ```go run main.go | go build main.go```
    * server starts 
      * [client web app](http://localhost:8080/client) located at http://localhost:8080/client
  * The example provided is a simple webapp that allows for the user to input:
    * tunnel name - the name identifier of the tunnel you would like to open
    * host        - local server addr that you would like to expose with ngrok
    * port        - port of the local server
    * protocol    - 0 - HTTP | 1 - TCP | 2 - TLS

## About

* An iOS app I was working on allows for the user to host a web server from their iOS device that acts as a simple web/chat server.
* I wanted to figure out a way to allow others that are outside of the local network to connect and chat if the device's network is protected by a strict firewall.
* I've used ngrok for a few other projects and figured it would be a decent place to start to test out the capabilities of whether or not web servers hosted on iOS devices on a local network could be exposed to the World Wide Web, which ended up being easy to test and implement correctly.

### PkgGoDev & GoDoc

[![PkgGoDev](https://pkg.go.dev/badge/github.com/revzim/gongrok)](https://pkg.go.dev/github.com/revzim/gongrok)
[![GoDoc](https://godoc.org/github.com/revzim/gongrok?status.svg)](https://godoc.org/github.com/revzim/gongrok)

### Workflow
```
  * INIT CLIENT
  
  * INIT TUNNEL(S)

  * RUN NGROK BINARY W/ OPTIONS

  * ADD TUNNEL(S) TO CLIENT

  * CONNECT ALL CLIENT SERVER TUNNELS

  * TUNNELS STAY OPEN UNTIL CLOSE

```

## Author
  * revzim

#### Credits

inspired by:
  * [Node NGROK Wrapper](https://github.com/bubenshchykov/ngrok)
  * [gonnel](https://github.com/afdalwahyu/gonnel)
