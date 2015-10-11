/* Plugin provides all the required interfaces to implement a GoPlugin
 *
 * Available API :
 *
 * func PluginInit(pluginImplConf PluginImplConf) (*Plugin, error)
 * >> Initialize a Plugin with specified Configuration
 *
 * func (plugin *Plugin) RegisterMethod(funcName string, method func([]byte) []byte)
 * >> Register a method to be executed for a Specified Path
 *
 * func (plugin *Plugin) Start() error
 * >> Start the execution of the specifiec Plugin
 *
 * func (plugin *Plugin) Stop() error
 * >> Stop the execution of that Plugin.
 */

package GoPlug

import (
	"encoding/json"
	"fmt"
	log "github.com/spf13/jwalterweatherman"
	PluginConn "github.com/swarvanusg/GoPlug/PluginConn"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

/* The plugin impleentaion configuration  */
type PluginImplConf struct {
	// Plugin location path
	PluginLoc string
	// The Name of the plugin
	Name string
	// The namespace of the plugin [optional - default: nil]
	Namespace string
	// The URL to reach the plugin over http (ie unix://ExamplePlug) [optional - default: unix://<Namespace><Name>]
	Url string
	// The LazyLoad configuration [optional - default: false]
	LazyLoad bool
	// The Function that would be called on Plugin Activation
	Activator func([]byte) []byte
	// The Function that would be called on Plugin DeActivation
	Stopper func([]byte) []byte
}

type PluginImpl struct {
	pluginServer   *PluginConn.PluginServer
	methodRegistry map[string]func([]byte) []byte
	sockFile       string
	addr           string
	confFile       string
}

// channel list per callback
var channelMap map[string]chan []byte

/* Init a plugin for a specific Plugin Conf */
func PluginInit(pluginImplConf PluginImplConf) (*PluginImpl, error) {

	plugin := &PluginImpl{}
	pluginConf := PluginConf{}

	// Check pluginImplConf
	if pluginImplConf.PluginLoc == "" {
		return nil, fmt.Errorf("Invalid Configuration : PluginLoc file should be specified")
	}

	// Check name
	if pluginImplConf.Name == "" {
		return nil, fmt.Errorf("Invalid Configuration : Name should be specified")
	}
	pluginConf.Name = pluginImplConf.Name
	pluginConf.NameSpace = pluginImplConf.Namespace

	// Check url
	pluginConf.Url = pluginImplConf.Url
	if pluginImplConf.Url == "" {
		pluginConf.Url = "unix://" + pluginConf.NameSpace + pluginConf.Name
	}

	// Get conf file and Sock
	confFile := filepath.Join(pluginImplConf.PluginLoc, pluginConf.NameSpace+pluginConf.Name+".pconf")
	pwd, _ := os.Getwd()
	sockFileLoc := filepath.Join(pwd, pluginImplConf.PluginLoc)
	pluginConf.Sock = filepath.Join(sockFileLoc, pluginConf.NameSpace+pluginConf.Name+".sock")

	// Get Lazyload
	pluginConf.LazyLoad = pluginImplConf.LazyLoad

	// Load Plugin Configuration
	confSaveError := saveConfigs(confFile, pluginConf)
	if confSaveError != nil {
		//fmt.Println("Configuration save failed to the file: ", confFile, ", Error: ", confSaveError)
		return nil, fmt.Errorf("Failed to save Configuration in file")
	}
	plugin.sockFile = pluginConf.Sock
	plugin.addr = pluginConf.Url

	// Initiate the Method Registry
	plugin.methodRegistry = make(map[string]func([]byte) []byte)

	plugin.methodRegistry["Activate"] = pluginImplConf.Activator
	plugin.methodRegistry["Stop"] = pluginImplConf.Stopper
	plugin.methodRegistry["RegisterCallback"] = callbackExecute

	plugin.confFile = confFile

	return plugin, nil
}

/* Internal Method: To implement the callback */
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
	fmt.Printf("Put the channel in the channelmap for: %s \n", funcName)
	channelMap[funcName] = channel

	fmt.Println("Waiting for data from channel")
	// Wait for data from channel
	returnData := <-channel
	fmt.Println("Got data from channel: ", string(returnData))

	return returnData
}

/* Internal Method: To Register method for the Plugin */
func (plugin *PluginImpl) Register() {

	http.Handle("/", plugin)
}

/* Internal Method: To handle all http request */
func (plugin *PluginImpl) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	methodName := strings.Split(req.URL.Path, "/")[1]
	//fmt.Printf("URL found: %s\n", methodName)
	if methodName == "" {
		res.WriteHeader(400)
	} else {
		method, ok := plugin.methodRegistry[methodName]
		if ok {
			// Check if the method is Activate
			if methodName == "Activate" {
				methodReg := plugin.methodRegistry
				methods := make([]string, len(methodReg))
				idx := 0
				for key, _ := range methodReg {
					methods[idx] = key
					idx++
				}
				data, marshalErr := json.Marshal(methods)
				if marshalErr != nil {
					//fmt.Println("failed to marshal methods")
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

/* Method to register function for the plugin */
func (plugin *PluginImpl) RegisterMethod(method func([]byte) []byte) {
	// Get the name of the method
	funcName := GetFuncName(method)
	plugin.methodRegistry[funcName] = method
}

/* Method to notify a registered callback */
func (plugin *PluginImpl) Notify(callBack string, data []byte) error {

	fmt.Println("Notifying the callback: ", callBack)
	// Pnthread : on getting the notifcation and user data it puts the data on the channel
	// Get the channel from global channel map
	channel, ok := channelMap[callBack]
	if !ok {
		return fmt.Errorf("Callback could not be found for: %s", callBack)
	}
	fmt.Println("Sending data to channel")
	fmt.Println("Sending data to channel")
	// Send the data to the channel
	channel <- data
	fmt.Println("data sent on channel")

	return nil
}

/* Start the Plugin Service */
func (plugin *PluginImpl) Start() error {

	sockFile := plugin.sockFile
	addr := plugin.addr
	// Create the Plugin Server
	config := &PluginConn.ServerConfiguration{Registrar: plugin, SockFile: sockFile, Addr: addr}
	server, err := PluginConn.NewPluginServer(config)
	if err != nil {
		//fmt.Printf("Failed to Create server\n")
		return fmt.Errorf("Failed to Create the server")
	}
	plugin.pluginServer = server

	plugin.pluginServer.Start()

	return nil
}

/* Stop the Plugin service */
func (plugin *PluginImpl) Stop() error {
	err := plugin.pluginServer.Shutdown()
	if err != nil {
		return err
	}
	err = os.Remove(plugin.confFile)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	channelMap = make(map[string]chan []byte)
}
