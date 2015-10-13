/* PluginReg is the Plugin Registry That keeps track of the Plugin
 * Which are Discovered, Loaded, and Activated
 */

package GoPlug

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/spf13/jwalterweatherman"
	PluginConn "github.com/swarvanusg/GoPlug/PluginConn"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
)

var (
	// An error to indicate the plugin not discovered
	ConfigLoadFailed = errors.New("Configuration load failed")

	// An error to indicate the plugin not discovered
	PluginNotDiscovered = errors.New("Plugin is not Discovered")

	// An error to indicate the connection to plugin could not be made
	PluginConnFailed = errors.New("Plugin to connection failed")

	// An error to indicate the plugin is already loaded
	PluginLoaded = errors.New("Plugin is already loaded")

	// Conf Extension
	DefaultConfExt = ".pconf"

	// Default Interval for Discovery search in MS
	DefaultInterval = 1000
)

type Plugin struct {
	// The name of the Plugin
	PluginName string
	// The nameSpace of the Plugin
	PluginNameSpace string
	// The URL to reach the Plugin
	PluginUrl string
	// The plugin Connection
	pluginConn *PluginConn.PluginClient
	// The plugin supported function
	methods []string
	// The plugin registered callback
	callbacks map[string]bool
	// Plugin disconnected state
	connected bool
}

/* PluginRegConf provides the configuration to create a plugin registry
 */
type PluginRegConf struct {
	// The location to search for Plugin. Default is .
	PluginLocation string
	// To enable and disable autoDisover. For auto Loading autoDiscover
	// Should be enabled. Default is false
	AutoDiscover bool
	// The conf Extension. Default is .conf
	ConfExt string
}

/* PluginReg should be created per types of Plugin
 * each PluginReg monitor a specific location */
type PluginReg struct {
	// The discoveredPlugin list
	DiscoveredPlugin map[string]PluginConf
	// The Loaded Plugin list
	PluginReg map[string]*Plugin
	// The waitgroup to wait for till PluginRegistry doesn't stop
	Wg *sync.WaitGroup
	// The Plugin search location
	PluginLocation string
	// The mutex to sync the Plugin reg access
	RegAccess *sync.Mutex
	// The flag to stop PluginRegistry Service
	StopFlag bool
	// The Conf Extension
	ConfExt string
	// To enable and disable autoDisover. For auto Loading autoDiscover
	// Should be enabled. Default is false
	AutoDiscover bool
}

/* Function is called to inititate the PluginRegistry as per the Plugin registry Configuration.
   It initiate and return a plugin registry pointer that could be used to manage plugins.
   If Discovery is enabled the DiscoverService Starts */
func PluginRegInit(regConf PluginRegConf) (*PluginReg, error) {

	var wg sync.WaitGroup

	pluginLocation := regConf.PluginLocation
	autoDiscover := regConf.AutoDiscover
	confExt := regConf.ConfExt
	if confExt == "" {
		confExt = DefaultConfExt
	}

	pluginReg := &PluginReg{}
	pluginReg.DiscoveredPlugin = make(map[string]PluginConf)
	pluginReg.PluginReg = make(map[string]*Plugin)
	pluginReg.PluginLocation = pluginLocation
	pluginReg.Wg = &wg
	pluginReg.RegAccess = &sync.Mutex{}
	pluginReg.StopFlag = false
	pluginReg.ConfExt = confExt
	pluginReg.AutoDiscover = autoDiscover
	// Check if autoDiscovery Is enabled
	if autoDiscover == true {
		// Start Plugin Discovery Service
		wg.Add(1)
		go discoverPlugin(&wg, pluginReg)
	}
	return pluginReg, nil
}

/* Function to wait for PluginReg Discovery service to be stopped. If its not started then it return immediately */
func (pluginReg *PluginReg) WaitForStop() {
	pluginReg.Wg.Wait()
}

/* Function to stop the Plugin Registry service. It stops the discovery service */
func (pluginReg *PluginReg) Stop() {
	pluginReg.StopFlag = true
}

/* Function for the routine to discover services */
func discoverPlugin(wg *sync.WaitGroup, pluginReg *PluginReg) {
	defer wg.Done()
	/* loop to Check for the Plugin Update */
	for true {
		pluginLocation := pluginReg.PluginLocation
		// Check the plugin location for a new plugin
		files, dirReadError := ioutil.ReadDir(pluginLocation)
		if dirReadError != nil {
			break
		}
		// Check for range of files in the location
		for _, f := range files {
			var name string
			name = f.Name()
			if f.IsDir() {
				// Skip if it is a directory */
				continue
			}
			ext := filepath.Ext(name)
			// Check if it is a Configuration File
			if ext == pluginReg.ConfExt {
				// Load new plugin Conf
				confFile := filepath.Join(pluginLocation, name)
				pluginConf, confLoadError := loadConfigs(confFile)
				if confLoadError != nil {
					log.ERROR.Println("Configuration load failed for file: ", confFile, ", Error: ", confLoadError)
					continue
				}
				appPlugin := pluginConf.NameSpace + pluginConf.Name
				// Check if in the discovered plugin list
				_, ok := pluginReg.DiscoveredPlugin[appPlugin]
				if !ok {
					// Store the config in the AppPlugin
					pluginReg.DiscoveredPlugin[appPlugin] = pluginConf

					// Check the lazyLoad conf.
					// if lazy load is disabled. Load it
					if pluginConf.LazyLoad == false {
						_, loadErr := pluginReg.LoadPlugin(pluginConf.Name, pluginConf.NameSpace)
						if loadErr != nil {
							log.ERROR.Println("Plugin load failed: ", loadErr)
						}
					}
				}
			}
		}
		// Check if stop file has been raised
		if pluginReg.StopFlag {
			break
		}
		// Wait for 1 sec
		time.Sleep(time.Duration(DefaultInterval))
	}
}

/* Check if a plugin is discovered by the plugin registry discovery service automatically or is discover implicitly */
func (pluginReg *PluginReg) IsDiscovered(pluginname string, namespace string) bool {

	appplugin := namespace + pluginname
	//pluginreg.regaccess.lock()
	//defer pluginreg.regaccess.unlock()
	return pluginReg.isDiscovered(appplugin)
}

/* Internal: Check if a plugin is already discovered */
func (pluginReg *PluginReg) isDiscovered(appPlugin string) bool {
	_, pluginDiscovered := pluginReg.DiscoveredPlugin[appPlugin]
	if !pluginDiscovered {
		return false
	}
	return true
}

/* Check if a plugin is loaded and active by the plugin registry automatically or is loaded implicitly */
func (pluginReg *PluginReg) IsLoaded(pluginname string, namespace string) bool {

	appplugin := namespace + pluginname
	pluginReg.RegAccess.Lock()
	defer pluginReg.RegAccess.Unlock()
	return pluginReg.isLoaded(appplugin)
}

/* Internal: Check if a plugin is already loaded and active */
func (pluginReg *PluginReg) isLoaded(appPlugin string) bool {

	plugin := pluginReg.getPlugin(appPlugin)
	if plugin == nil {
		return false
	}
	return true
}

/* Unload a Plugin from the plugin Registry. It invokes a stop request to the plugin.
   (It doesn't remove the Plugin from Discovered Plugin List) */
func (pluginReg *PluginReg) UnloadPlugin(plugin *Plugin) error {

	// Generate Discovered plugin name
	appPlugin := plugin.PluginNameSpace + plugin.PluginName

	// Initiate Locking
	pluginReg.RegAccess.Lock()
	defer pluginReg.RegAccess.Unlock()

	// Send the Stop request
	stopErr := plugin.stop()
	if stopErr != nil {
		log.ERROR.Println("Failed to stop the plugin")
	}

	// Close the connection
	plugin.pluginConn.Close()

	// Delete the plugin from registry
	delete(pluginReg.PluginReg, appPlugin)

	return nil
}

/* Load the plugin to the plugin Registry explicitly when lazy load is active.
(if The discovery Process is not running, It search for the plugin and then load it to the registry)
*/
func (pluginReg *PluginReg) LoadPlugin(pluginName string, pluginNamespace string) (*Plugin, error) {

	var conf PluginConf

	// Generate Discovered plugin name
	appPlugin := pluginNamespace + pluginName

	if pluginReg.AutoDiscover == false {
		var confLoadError error
		pluginLocation := pluginReg.PluginLocation
		// Load new plugin Conf
		confFile := filepath.Join(pluginLocation, pluginName+pluginReg.ConfExt)
		conf, confLoadError = loadConfigs(confFile)
		if confLoadError != nil {
			return nil, ConfigLoadFailed
		}

		// Store the config in the AppPlugin
		pluginReg.DiscoveredPlugin[appPlugin] = conf

	} else {
		var discovered bool
		// Check if Plugin is already discovered
		conf, discovered = pluginReg.DiscoveredPlugin[appPlugin]
		if !discovered {
			return nil, PluginNotDiscovered
		}
	}

	// Initiate Locking
	pluginReg.RegAccess.Lock()
	defer pluginReg.RegAccess.Unlock()

	// Check if Plugin is already loaded
	if pluginReg.isLoaded(appPlugin) {
		return nil, PluginLoaded
	}

	sockFile := conf.Sock

	// Initiate Connection to a Plugin
	pluginConn, connErr := PluginConn.NewPluginClient(sockFile)
	if connErr != nil {
		return nil, PluginConnFailed
	}

	plugin := &Plugin{}
	plugin.PluginName = conf.Name
	plugin.PluginNameSpace = conf.NameSpace
	plugin.PluginUrl = conf.Url
	plugin.pluginConn = pluginConn
	plugin.callbacks = make(map[string]bool)
	pluginReg.PluginReg[appPlugin] = plugin

	// Activate the plugin
	activateErr := plugin.activate()
	if activateErr != nil {
		return plugin, activateErr
	}

	return plugin, nil
}

/* Get a plugin from the Plugin registry for a specified name and namespace */
func (pluginReg *PluginReg) GetPlugin(pluginName string, pluginNamespace string) *Plugin {

	appPlugin := pluginNamespace + pluginName
	pluginReg.RegAccess.Lock()
	defer pluginReg.RegAccess.Unlock()
	return pluginReg.getPlugin(appPlugin)
}

func (pluginReg *PluginReg) getPlugin(appPlugin string) *Plugin {
	/* Check if the plugin is Loaded in the plugin map */
	plugin, pluginFound := pluginReg.PluginReg[appPlugin]
	if pluginFound {
		return plugin
	}
	return nil
}

// function to check plugin status
func (plugin *Plugin) checkConnection() bool {
	if plugin.Ping() != nil {
		return false
	}
	return true
}

// Activate a plugin
func (plugin *Plugin) activate() error {
	pluginUrl := plugin.PluginUrl
	pluginConn := plugin.pluginConn

	requestUrl := pluginUrl + "/Activate"
	request := &PluginConn.PluginRequest{Url: requestUrl, Body: nil}

	resp, reqerr := pluginConn.Request(request)
	if reqerr != nil {
		plugin.connected = false
		return reqerr
	}
	if resp.Status != "200 OK" {
		return fmt.Errorf("request failed. Status: %s", resp.Status)
	}

	// Get the response
	unmarshalError := json.Unmarshal(resp.Body, &plugin.methods)
	if unmarshalError != nil {
		return fmt.Errorf("Json Unmarshal failed: %s", unmarshalError)
	}

	return nil
}

// Deactivate a plugin
func (plugin *Plugin) stop() error {
	pluginUrl := plugin.PluginUrl
	pluginConn := plugin.pluginConn

	requestUrl := pluginUrl + "/Stop"
	request := &PluginConn.PluginRequest{Url: requestUrl, Body: nil}

	resp, err := pluginConn.Request(request)
	if err != nil {
		return err
	}
	if resp.Status != "200 OK" {
		return fmt.Errorf("request failed")
	}

	return nil
}

/* Get the list of available (registered) methods for a specific plugin */
func (plugin *Plugin) GetMethods() []string {

	var methods []string
	methods = plugin.methods
	return methods
}

/* Register a callback that will be called on notification from the plugin */
func (plugin *Plugin) RegisterCallback(function func([]byte)) error {
	funcName := getFuncName(function)
	if funcName == "" {
		return fmt.Errorf("Failed to get the method name")
	}
	// Check if the callback is already registered
	_, ok := plugin.callbacks[funcName]
	if ok {
		return fmt.Errorf("The callback is already Registerd")
	}
	// Put the callback function in the callbacks map
	plugin.callbacks[funcName] = false

	// Start the execution thread
	go plugin.executeCallback(funcName, function)

	return nil
}

// Internal:  thread body to execute a callback request
func (plugin *Plugin) executeCallback(funcName string, function func([]byte)) {
	// wrap the method name in bytes
	data, marshalErr := json.Marshal(funcName)
	if marshalErr != nil {
		log.ERROR.Printf("Json Marshal Failed to encode method name")
		return
	}

	pluginUrl := plugin.PluginUrl
	pluginConn := plugin.pluginConn

	requestUrl := pluginUrl + "/" + "RegisterCallback"
	request := &PluginConn.PluginRequest{Url: requestUrl, Body: data}

	//	for plugin.callbacks[funcName] == false {
	for true {
		resp, err := pluginConn.Request(request)
		if err != nil {
			plugin.connected = false
			log.FATAL.Fatalf("Failed to sent CallBack Execution Request: %v", err)
			return
		}
		if resp.Status != "200 OK" {
			log.FATAL.Fatalf("Failed to sent callback request")
			return
		}
		// get the data from resp
		callBackInput := resp.Body
		// call the callback
		function(callBackInput)
	}
}

/* Executes a specific plugin method by the method name. Each method takes a byte array as input
   and returns a byte array as output */
func (plugin *Plugin) Execute(funcName string, body []byte) (error, []byte) {

	pluginUrl := plugin.PluginUrl
	pluginConn := plugin.pluginConn

	requestUrl := pluginUrl + "/" + funcName
	request := &PluginConn.PluginRequest{Url: requestUrl, Body: body}

	resp, err := pluginConn.Request(request)
	if err != nil {
		plugin.connected = false
		return err, nil
	}
	if resp.Status != "200 OK" {
		return fmt.Errorf("request failed"), nil
	}

	return nil, resp.Body
}

/* Ping a specific plugin to check the plugin status */
func (plugin *Plugin) Ping() error {

	pluginUrl := plugin.PluginUrl
	pluginConn := plugin.pluginConn

	testData := "Test Data"
	sendData := []byte(testData)

	requestUrl := pluginUrl + "/" + "Ping"
	request := &PluginConn.PluginRequest{Url: requestUrl, Body: sendData}

	resp, err := pluginConn.Request(request)
	if err != nil {
		plugin.connected = false
		return err
	}
	if resp.Status != "200 OK" {
		return fmt.Errorf("request failed")
	}

	receivedData := string(resp.Body)

	if receivedData != testData {
		return fmt.Errorf("Received data is different than sent one")
	}

	return nil
}
