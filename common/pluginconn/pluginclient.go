package pluginconn

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	httputil "net/http/httputil"
)

type PluginClient struct {
	Conn *httputil.ClientConn
}

type PluginRequest struct {
	Url  string
	Body []byte
}

type PluginResponse struct {
	Status string
	Body   []byte
}

func NewPluginClient(sockFile string) (*PluginClient, error) {

	conn, connErr := net.Dial("unix", sockFile)
	if connErr != nil {
		fmt.Printf("Connection could not be initiated")
		return nil, connErr
	}

	/* Use default buffer */
	clientConn := httputil.NewClientConn(conn, nil)
	if clientConn == nil {
		fmt.Printf("Connection Could not be Configured")
		return nil, connErr
	}
	pluginConn := &PluginClient{clientConn}

	return pluginConn, nil
}

func (pluginConn *PluginClient) Request(request *PluginRequest) (*PluginResponse, error) {

	var url = request.Url
	var req *http.Request
	var newReqErr error

	if request.Body != nil {
		var jsonStr = request.Body
		req, newReqErr = http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		if newReqErr != nil {
			fmt.Printf("Request Could not be prepared")
			return nil, newReqErr
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, newReqErr = http.NewRequest("POST", url, nil)
		if newReqErr != nil {
			fmt.Printf("Request Could not be prepared")
			return nil, newReqErr
		}
	}
	conn := pluginConn.Conn

	resp, reqErr := conn.Do(req)
	if reqErr != nil {
		fmt.Printf("Request Could not be send")
		return nil, reqErr
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	if body != nil {
		fmt.Println("response Body:", string(body))
	}

	response := &PluginResponse{}
	response.Status = resp.Status
	response.Body = body

	return response, nil
}

func (pluginConn *PluginClient) Close() error {

	pluginConn.Conn.Close()
	return nil
}
