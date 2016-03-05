/* Plugin provides all the required interfaces to implement a GoPlugin
 */

package pluginlib

import (
	"encoding/json"
	"fmt"
	log "github.com/spf13/jwalterweatherman"
	plugreg "github.com/swarvanusg/GoPlug"
	common "github.com/swarvanusg/GoPlug/common"
	PluginConn "github.com/swarvanusg/GoPlug/common/pluginconn"
	"io/ioutil"
	"net/http"
	"strings"
)

type Plugintype interface {
	// Called to init your plugin when it is loaded
	Init() error
	// Called to start your plugin
	Start(map[string]interface{}) error
	// Called to stop your plugin
	Stop() error
}

/* The Plugin Implentaion Struct to represent a Plugin, provides all the methods to be implemented */
type Plugin struct {
	pluginServer   *PluginConn.PluginServer
	methodRegistry []string
	methodObject   interface{}
	conf           *common.RuntimeConf
	started        bool
}

// channel list per callback that are registered
var channelMap map[string]chan []byte

/* Initialize a plugin as per the provided plugin implementation configuration.
   It returns a pointer to a Plugin that is used to perfom different operation
   on the implementde plugin */
func PluginInit(plugin Plugintype) (*Plugin, error) {

	var plugin = &Plugin{}

	// Load the plugin runtime conf from pluginPath
	pluginConf, err := common.LoadRuntimeConfigs(DefaultPluginConfFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to load the config file")
	}

	plugin.methodRegistry = make(map[string]func([]byte) []byte)
	channelMap = make(map[string]chan []byte)

	// Register the basic method
	pluginType := reflect.TypeOf(plugin)
	for i := 0; i < fooType.NumMethod(); i++ {
		plugin.methodRegistry = append(plugin.methodRegistry, fooType.Method(i))
	}

	// Register plugin object
	plugin.methodObject = plugin

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
func (plugin *Plugin) Register() {
	http.Handle("/", plugin)
}

/* Internal Method: Executes a method after unwrapping its arguments */
func executeMethod(data interface{}, name string, data []byte) []byte {
	// The receiver reades the Json and Do the magic
	json_data := common.ReadJson(data)

	argsspace := make([]reflect.Value, 0)
	for _, arg := range json_data {
		argsspace = append(argsspace, reflect.ValueOf(arg))
	}
	values := reflect.ValueOf(data).MethodByName(name).Call(argsspace)
	return_vals := make([]interface{}, 0)
	for _, value := range values {
		return_vals = append(return_vals, value.Interface())
	}
	returnbytes = common.CreateJson(return_vals...)
	return returnbytes
}

/* Internal Method: Default handler to serve all http request that comes to the plugin. Should not be called explicitly */
func (plugin *Plugin) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	methodName := strings.Split(req.URL.Path, "/")[1]
	if methodName == "" {
		res.WriteHeader(400)
	} else {
		methods := plugin.methodRegistry
		ok := false
		for _, method := range methods {
			if method == methodname {
				ok = true
				break
			}
		}
		if ok {
			// Check if the method is Activate
			if methodName == "Start" {
				methodReg := plugin.methodRegistry
				var methods []string = nil
				for _, method := range methodReg {
					switch method {
					case "Start":
					case "Stop":
					case "Init":
					default:
						methods = append(methods, method)
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
			input, _ := ioutil.ReadAll(req.Body)
			returnData := executeMethod(plugin.methodObject, methodName, input)
			if returnData != nil {
				res.Write(returnData)
			}
			res.WriteHeader(200)
		} else {
			res.WriteHeader(400)
		}
	}
}

/* Method to notify a callback registered by the application by the name of the callback.
   User could sent input bytes for the callback. Callback doesn't return anything */
func (plugin *Plugin) Notify(callBack string, args ...interface{}) error {

	data := CreateJson(args...)

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
func (plugin *Plugin) Start() error {

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
func (plugin *Plugin) Stop() error {
	err := plugin.pluginServer.Shutdown()
	if err != nil {
		return err
	}
	return nil
}
