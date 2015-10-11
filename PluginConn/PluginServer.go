package PluginConn

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

// interface for http handler registrations
type HttpHandlerRegistrar interface {
	Register()
}

// HTTPServer is used to wrap an Agent and expose various API's
// in a RESTful manner
type PluginServer struct {
	Mux      *http.ServeMux
	Listener net.Listener
	Addr     string
}

// configuration for the http server
type ServerConfiguration struct {
	Registrar HttpHandlerRegistrar
	SockFile  string
	Addr      string
}

// Create a new HTTP Server
func NewPluginServer(config *ServerConfiguration) (*PluginServer, error) {

	// Get listener for the Http server
	listener, err := net.Listen("unix", config.SockFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to set Listner: %s", err)
	}

	// Create the mux
	//mux := http.NewServeMux()

	// Create the server
	server := &PluginServer{
		Mux:      nil,
		Listener: listener,
		Addr:     config.Addr,
	}

	// register the http handlers
	config.Registrar.Register()

	return server, nil
}

// Start the http Server
func (s *PluginServer) Start() {

	/* Each request is served in a separate thread */
	go http.Serve(s.Listener, nil)

}

// Shutdown is used to shutdown the HTTP server
func (s *PluginServer) Shutdown() error {
	if s != nil {
		fmt.Printf("[DEBUG] http: Shutting down http server (%v)\n", s.Addr)
		err := s.Listener.Close()
		if err != nil {
			fmt.Printf("[ERROR] Failed to close http listener: %v\n", err)
			return err
		}
	}
	return nil
}

// Write a json string with given header code
func WriteJsonResponse(v interface{}, code int, w http.ResponseWriter) error {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return nil
}
