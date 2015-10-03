/* Plugin provides all the required interfaces to implement a GoPlugin
 *
 * Available API :
 *
 * func PluginInit(pluginImplConf PluginImplConf) (*Plugin, error)
 * >> Initialize a Plugin with specified Configuration
 *
 * func (plugin *Plugin) RegisterFunc(funcName string, method func([]byte) []byte)
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
	"com.ss/goplugin/PluginConn"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type PluginImplConf struct {
	ConfFile  string
	Activator func([]byte) []byte
	Stopper   func([]byte) []byte
}

type PluginImpl struct {
	pluginServer   *PluginConn.PluginServer
	methodRegistry map[string]func([]byte) []byte
	sockFile       string
	addr           string
}

/* Init a plugin for a specific Plugin Conf */
func PluginInit(pluginImplConf PluginImplConf) (*PluginImpl, error) {

	plugin := &PluginImpl{}

	// Load Plugin Configuration
	pluginConf, confLoadError := loadConfigs(pluginImplConf.ConfFile)
	if confLoadError != nil {
		fmt.Println("Configuration load failed for file: ", pluginImplConf.ConfFile, ", Error: ", confLoadError)
		return nil, fmt.Errorf("Failed to load Configuration")
	}
	plugin.sockFile = pluginConf.Sock
	plugin.addr = pluginConf.Url

	// Load Plugin value
	/*
		fmt.Printf("Name: %s\n", pluginConf.Name)
		fmt.Printf("NameSpace: %s\n", pluginConf.NameSpace)
		fmt.Printf("Url: %s\n", pluginConf.Url)
		fmt.Printf("Sock: %s\n", pluginConf.Sock)
		fmt.Printf("LazyLoad: %d\n", pluginConf.LazyLoad)
	*/

	// Initiate the Method Registry
	plugin.methodRegistry = make(map[string]func([]byte) []byte)

	plugin.methodRegistry["Activate"] = pluginImplConf.Activator
	plugin.methodRegistry["Stop"] = pluginImplConf.Stopper

	return plugin, nil
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
func (plugin *PluginImpl) RegisterFunc(funcName string, method func([]byte) []byte) {
	plugin.methodRegistry[funcName] = method
}

/* Start the Plugin Service */
func (plugin *PluginImpl) Start() error {

	sockFile := plugin.sockFile
	addr := plugin.addr
	// Create the Plugin Server
	config := &PluginConn.ServerConfiguration{Registrar: plugin, SockFile: sockFile, Addr: addr}
	server, err := PluginConn.NewPluginServer(config)
	if err != nil {
		fmt.Printf("Failed to Create server\n")
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
	return nil
}
