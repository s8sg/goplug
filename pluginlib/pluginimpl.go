/* Plugin provides all the required interfaces to implement a GoPlugin
 */

package pluginlib

import (
	"encoding/json"
	"fmt"
	log "github.com/spf13/jwalterweatherman"
	. "github.com/swarvanusg/GoPlug"
	PluginConn "github.com/swarvanusg/GoPlug/pluginconn"
	"io/ioutil"
	"net/http"
	"strings"
)

type Plugintype interface {
	// Called to init your plugin when it is loaded
	Init()
	// Called to start your plugin
	Start([]byte) []byte
	// Called to stop your plugin
	Stop([]byte) []byte
}

/* The Plugin Implentaion Struct to represent a Plugin, provides all the methods to be implemented */
type PluginImpl struct {
	pluginServer   *PluginConn.PluginServer
	methodRegistry map[string]func([]byte) []byte
	conf           *RuntimeConf
	started        bool
}

// channel list per callback that are registered
var channelMap map[string]chan []byte

/* Initialize a plugin as per the provided plugin implementation configuration.
   It returns a pointer to a PluginImpl that is used to perfom different operation
   on the implementde plugin */
func PluginInit(plugin Plugintype) (*PluginImpl, error) {

	var plugin = &PluginImpl{}

	// Load the plugin runtime conf from pluginPath
	pluginConf, err := loadRuntimeConfigs(DefaultPluginConfFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to load the config file")
	}

	plugin.methodRegistry = make(map[string]func([]byte) []byte)
	channelMap = make(map[string]chan []byte)

	// Register the basic method
	plugin.methodRegistry["Start"] = plugin.Start
	plugin.methodRegistry["Stop"] = plugin.Stop
	plugin.methodRegistry["RegisterCallback"] = callbackExecute
	plugin.methodRegistry["Ping"] = ping

	plugin.conf = &pluginConf

	return plugin, nil
}

/* Internal Method: To ping a plugin */
func ping(data []byte) []byte {
	// Return the same data
	return data
}

/* Internal Method: To execute a callback -- wait for a data in a channel to be notified */
func callbackExecute(data []byte) []byte {

	// get the function name
	var funcName string
	err := json.Unmarshal(data, &funcName)
	if err != nil {
		log.FATAL.Fatalf("Failed to get the func name: %v", err)
		return nil
	}

	// Create a new channel
	channel := make(chan []byte, 0)

	// Put the channel in the channelmap
	channelMap[funcName] = channel

	// Wait for data from channel
	returnData := <-channel

	return returnData
}

/* Internal Method: Used to register a handle method for the incoming request to plugin. Should not be called explicitly */
func (plugin *PluginImpl) Register() {
	http.Handle("/", plugin)
}

/* Internal Method: Default handler to serve all http request that comes to the plugin. Should not be called explicitly */
func (plugin *PluginImpl) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	methodName := strings.Split(req.URL.Path, "/")[1]
	if methodName == "" {
		res.WriteHeader(400)
	} else {
		method, ok := plugin.methodRegistry[methodName]
		if ok {
			// Check if the method is Activate
			if methodName == "Start" {
				methodReg := plugin.methodRegistry
				methods := make([]string, len(methodReg))
				idx := 0
				// get all the register method
				for key, _ := range methodReg {
					// Skip the implicit functions
					if key != "Start" && key != "Stop" && key != "RegisterCallback" && key != "Ping" {
						methods[idx] = key
						idx++
					}
				}
				// marshal the method list sent on activation
				data, marshalErr := json.Marshal(methods)
				if marshalErr != nil {
					res.WriteHeader(400)
				}
				// Write the methods list
				res.Write(data)
			}

			defer req.Body.Close()
			body, _ := ioutil.ReadAll(req.Body)
			returnData := method(body)
			if returnData != nil {
				res.Write(returnData)
			}
			res.WriteHeader(200)
		} else {
			res.WriteHeader(400)
		}
	}
}

/* Method to register a function for the plugin that could be invoked by the application.
   Function Prototype: func ([]byte) []byte */
func (plugin *PluginImpl) RegisterMethod(method func([]byte) []byte) error {
	if plugin.started {
		return fmt.Errorf("Method can't be registered once plugin has started")
	}
	// Get the name of the method
	funcName := getFuncName(method)
	plugin.methodRegistry[funcName] = method
}

/* Method to notify a callback registered by the application by the name of the callback.
   User could sent input bytes for the callback. Callback doesn't return anything */
func (plugin *PluginImpl) Notify(callBack string, data []byte) error {

	// Pnthread : on getting the notifcation and user data it puts the data on the channel
	// Get the channel from global channel map
	channel, ok := channelMap[callBack]
	if !ok {
		return fmt.Errorf("Callback could not be found for: %s", callBack)
	}
	// Send the data to the channel
	channel <- data

	return nil
}

/* Used to start the Plugin Service. It makes a plugin operable and discoverable by application */
func (plugin *PluginImpl) Start() error {

	sockFile := plugin.conf.Sock
	addr := plugin.conf.Url

	// Create the Plugin Server
	config := &PluginConn.ServerConfiguration{Registrar: plugin, SockFile: sockFile, Addr: addr}
	server, err := PluginConn.NewPluginServer(config)
	if err != nil {
		return fmt.Errorf("Failed to Create the server")
	}
	plugin.pluginServer = server

	// Start the server (it will add the sock file in proper position)
	plugin.pluginServer.Start()

	// Set the plugin start flag
	plugin.started = true

	return nil
}

/* Used to stop the Plugin service. It makes the plugin hidden from the application and stops all functionalities */
func (plugin *PluginImpl) Stop() error {
	err := plugin.pluginServer.Shutdown()
	if err != nil {
		return err
	}
	return nil
}
