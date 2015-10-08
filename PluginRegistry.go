/* PluginReg is the Plugin Registry That keeps track of the Plugin
 * Which are Discovered, Loaded, and Activated
 *
 * Available API :
 *
 * func PluginRegInit(regConf PluginRegConf) (*PluginReg, error)
 * >> It Intialize a new plugin reg as per the PluginRegConf specified
 *
 * func (pluginReg *PluginReg) IsDiscovered(pluginname string, namespace string) bool
 * >> Check if a plugin is already discovered
 *
 * func (pluginReg *PluginReg) IsLoaded(pluginname string, namespace string) bool
 * >> Check if a plugin is already Loaded
 *
 * func (pluginReg *PluginReg) LoadPlugin(pluginName string, pluginNamespace string) (*Plugin, error)
 * >> Load and activate a plugin manually if it is discovered
 *
 * func (pluginReg *PluginReg) UnloadPlugin(plugin *Plugin) error
 * >> Unload and stop a plugin manually if it is loaded
 *
 * func (pluginReg *PluginReg) GetPlugin(pluginName string, pluginNamespace string) *Plugin
 * >> Function to Get a Plugin When it is Loaded
 *
 * func (plugin *Plugin) Execute(funcName string, body []byte) (error, []byte)
 * >> Execute a function with given data set
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

	// Monitor Type Plugin
	MonitorPluginT = "Monitor"

	// Manage Type Plugin
	ManagePluginT = "Manage"

	// Conf Extension
	DefaultConfExt = ".conf"

	// Default Interval for Discovery search
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

/* Function is called to inititate the PluginRegistry
 * It initiate and return a plugin registry.
 * If Discovery is enabled the DiscoverService Starts */
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

/* Function to wait for PluginReg Discovery service to be stopped */
func (pluginReg *PluginReg) WaitForStop() {
	pluginReg.Wg.Wait()
}

/* Function to stop the PluginReg Discovery service */
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
			//fmt.Printf("Invalid Directory: %s\n", pluginLocation)
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
		time.Sleep(time.Duration(DefaultInterval * 100))
	}
}

/* Check if a plugin is already discovered */
func (pluginReg *PluginReg) IsDiscovered(pluginname string, namespace string) bool {

	appplugin := namespace + pluginname
	fmt.Printf("going to take lock\n")
	//pluginreg.regaccess.lock()
	//defer pluginreg.regaccess.unlock()
	fmt.Printf("lock aquired\n")
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

/* Check if a plugin is already loaded and active */
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

/* unload the plugin from the plugin Registry
 * (It doesn't remove the Plugin from Discoveed List)
 */
func (pluginReg *PluginReg) UnloadPlugin(plugin *Plugin) error {

	// Generate Discovered plugin name
	appPlugin := plugin.PluginNameSpace + plugin.PluginName

	// Initiate Locking
	pluginReg.RegAccess.Lock()
	defer pluginReg.RegAccess.Unlock()

	// get the Plugin that is already loaded
	/*
		plugin := pluginReg.getPlugin(appPlugin)
		if plugin == nil {
			return fmt.Errorf("plugin is not Loaded")
		}
	*/

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

/* Load the plugin to the plugin Registry
 * (if The discovery Process is not running, It discover the plugin)
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
	pluginReg.PluginReg[appPlugin] = plugin

	// Activate the plugin
	activateErr := plugin.activate()
	if activateErr != nil {
		return plugin, activateErr
	}

	return plugin, nil
}

// Get a plugin from the Plugin registry
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

// Activate a plugin
func (plugin *Plugin) activate() error {
	pluginUrl := plugin.PluginUrl
	pluginConn := plugin.pluginConn

	requestUrl := pluginUrl + "/Activate"
	request := &PluginConn.PluginRequest{Url: requestUrl, Body: nil}

	resp, reqerr := pluginConn.Request(request)
	if reqerr != nil {
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
	fmt.Println("Methods: ", plugin.methods)

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

// Get the list of available method
func (plugin *Plugin) GetMethods() []string {

	var methods []string
	methods = plugin.methods
	return methods
}

// Executes a specific plugin method
func (plugin *Plugin) Execute(funcName string, body []byte) (error, []byte) {

	pluginUrl := plugin.PluginUrl
	pluginConn := plugin.pluginConn

	requestUrl := pluginUrl + "/" + funcName
	request := &PluginConn.PluginRequest{Url: requestUrl, Body: body}

	resp, err := pluginConn.Request(request)
	if err != nil {
		return err, nil
	}
	if resp.Status != "200 OK" {
		return fmt.Errorf("request failed"), nil
	}

	return nil, resp.Body
}
