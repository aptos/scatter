package main

/*
#include <stdlib.h>
*/
import "C"

import (
  "encoding/json"
  "io/ioutil"
  "net/http"
  "unsafe"
)

type Command struct {
  URIs []string `json:"uris"`
}

type Result map[string]*Response
type Response struct {
  Body   string `json:"body"`
  Status int    `json:"status"`
  Err    string `json:"error"`
  Uri    string `json:"uri"`
}

func makeRequest(uri string) *Response {
  resp, err := http.Get(uri)
  if err != nil {
    return &Response{Status: resp.StatusCode, Err: err.Error()}
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return &Response{Status: resp.StatusCode, Err: err.Error()}
  }

  return &Response{Uri: uri, Body: string(body), Status: resp.StatusCode}
}

//export scatter_request
func scatter_request(data *C.char) *C.char {
  cmd := Command{}
  err := json.Unmarshal([]byte(C.GoString(data)), &cmd)
  if err != nil {
    // return err.Error()
    err_str := C.CString(err.Error())
    defer C.free(unsafe.Pointer(err_str))
    return err_str
  }

  c := make(chan *Response)
  for _, uri := range cmd.URIs {
    go func(_uri string) {
      c <- makeRequest(_uri)
    }(uri)
  }

  result := Result{}
  for i := 0; i < len(cmd.URIs); i++ {
    resp := <-c
    result[resp.Uri] = resp
  }

  close(c)

  b, err := json.Marshal(result)
  if err != nil {
    // return "{}"
    empty_str := C.CString("{}")
    defer C.free(unsafe.Pointer(empty_str))
    return empty_str
  }

  // return string(b)
  result_str := C.CString(string(b))
  defer C.free(unsafe.Pointer(result_str))
  return result_str
}

func main() {}
