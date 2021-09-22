package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// RPCError represents an error that is used as a part of a JSON-RPC JsonResponse
// object.
type RPCError struct {
	Code       int    `json:"Code,omitempty"`
	Message    string `json:"Message,omitempty"`
	StackTrace string `json:"StackTrace"`
	Err        error  `json:"Err"`
}

// JsonRequest represents a JSON-RPC request.
type JsonRequest struct {
	JsonRPC string      `json:"Jsonrpc"`
	Method  string      `json:"Method"`
	Params  interface{} `json:"Params"`
	Id      interface{} `json:"Id"`
}

// JsonResponse represents a JSON-RPC response.
type JsonResponse struct {
	Id      *interface{}    `json:"Id"`
	Result  json.RawMessage `json:"Result"`
	Error   *RPCError       `json:"Error"`
	Params  interface{}     `json:"Params"`
	Method  string          `json:"Method"`
	JsonRPC string          `json:"Jsonrpc"`
}

// OldParseResponse parses a raw JSON-RPC response into a JsonResponse.
//
// Deprecated: use ParseResponse instead.
func OldParseResponse(respondInBytes []byte) (*JsonResponse, error) {
	var respond JsonResponse
	err := json.Unmarshal(respondInBytes, &respond)
	if err != nil {
		return nil, err
	}

	if respond.Error != nil {
		return nil, fmt.Errorf("RPC returns an error: %v", respond.Error)
	}

	return &respond, nil
}

// ParseResponse parses a JSON-RPC response to val.
func ParseResponse(respondInBytes []byte, val interface{}) error {
	var respond JsonResponse
	err := json.Unmarshal(respondInBytes, &respond)
	if err != nil {
		if len(respondInBytes) == 0 {
			return fmt.Errorf("RPC response is empty")
		}
		log.Printf("%v, %v\n", len(respondInBytes), string(respondInBytes))
		return fmt.Errorf("un-marshal RPC-response error: %v", err)
	}

	if respond.Error != nil {
		return fmt.Errorf("RPC returns an error: %v", respond.Error)
	}

	if val == nil {
		return nil
	}

	err = json.Unmarshal(respond.Result, val)
	if err != nil {
		return err
	}

	return nil
}

// CreateJsonRequest creates a new JsonRequest given the method and parameters.
func createJsonRequest(jsonRPC, method string, params []interface{}, id interface{}) *JsonRequest {
	request := new(JsonRequest)
	request.JsonRPC = jsonRPC
	request.Method = method
	request.Id = id
	request.Params = params

	return request
}

func SendQuery(method string, params []interface{}) ([]byte, error) {
	if params == nil {
		params = make([]interface{}, 0)
	}
	request := createJsonRequest("1.0", method, params, 1)

	query, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return sendPostRequestWithQuery(string(query))
}

func sendPostRequestWithQuery(query string) ([]byte, error) {
	var jsonStr = []byte(query)
	req, _ := http.NewRequest("POST", fullnodeURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("DoReq %v\n", err)
		return []byte{}, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%v", resp.Status)
	} else {
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.Printf("BodyClose %v\n", err)
			}
		}()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ReadAll %v\n", err)
			return []byte{}, err
		}
		return body, nil
	}
}
