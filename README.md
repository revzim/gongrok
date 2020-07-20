# gongrok

## golang [ngrok](https://ngrok.com/) wrapper
  * `Secure introspectable tunnels to local host`

## Install

 *  [ngrok](https://ngrok.com/download) binary is required for use
    *  download the binary
    *  place binary within your working directory
    *  ```default: ./ngrok_bin/```
```
go get github.com/revzim/gongrok
```

## Example
* [WEB APP EXAMPLE](https://github.com/revzim/gongrok/example/webapp)
  * place ngrok binary download in ngrok path ```./ngrok_bin/```
  * build or run
    * ```go run main.go | go build main.go```
    * server will start & client located at ```http://localhost:8080/client```
