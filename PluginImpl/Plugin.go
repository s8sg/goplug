package PluginImpl

import (
	"com.ss/goplugin/PluginConn"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type PluginImplConf struct {
	ConfFile  string
	Activator func([]byte) []byte
	Stopper   func([]byte) []byte
}

type Plugin struct {
	pluginServer   *PluginConn.PluginServer
	methodRegistry map[string]func([]byte) []byte
	sockFile       string
	addr           string
}

/* The configuration of the plugin */
type PluginConf struct {
	Name      string
	NameSpace string
	Url       string
	Sock      string
	LazyLoad  bool
}

/* Init a plugin for a specific Plugin Conf */
func PluginInit(pluginImplConf PluginImplConf) (*Plugin, error) {

	plugin := &Plugin{}

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
func (plugin *Plugin) Register() {

	http.Handle("/", plugin)
}

/* Internal Method: To handle all http request */
func (plugin *Plugin) ServeHTTP(res http.ResponseWriter, req *http.Request) {

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
func (plugin *Plugin) RegisterFunc(funcName string, method func([]byte) []byte) {
	plugin.methodRegistry[funcName] = method
}

/* Start the Plugin Service */
func (plugin *Plugin) Start() error {

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
func (plugin *Plugin) Stop() error {
	err := plugin.pluginServer.Shutdown()
	if err != nil {
		return err
	}
	return nil
}

// Load configuration from file
func loadConfigs(fname string) (PluginConf, error) {
	// open the config file
	configuration := PluginConf{}
	file, err := os.Open(fname)
	if err != nil {
		return configuration, err
	}
	// load the config from file
	decoder := json.NewDecoder(file)
	loaderr := decoder.Decode(&configuration)
	if loaderr != nil {
		return configuration, loaderr
	}

	return configuration, nil
}
