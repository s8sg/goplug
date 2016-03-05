/* PluginReg is the Plugin Registry That keeps track of the Plugin
 * Which are Discovered, Loaded, and Activated
 */

package pluginmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/spf13/jwalterweatherman"
	commom "github.com/swarvanusg/GoPlug/common"
	PluginConn "github.com/swarvanusg/GoPlug/common/pluginconn"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"syscall"
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

	UntarError    = errors.New("Failed to unload the Tar file")
	SaveConfError = errors.New("Failed to save the plugin conf")

	// Conf Extension
	DefaultPluginConfFile        = "plugin.conf"
	DefaultPluginRuntimeConfFile = "runtime.conf"
	DefaultDiscoveredPlugin      = "discoveredplugin"
	DefaultTarExt                = ".tar"
	PluginBinary                 = "pluginmain"
	PluginSockFile               = "pluginconn.sock"
	PluginUrl                    = "unix://plugin"
	// Default Interval for Discovery search in MS
	DefaultInterval = 500 * time.Millisecond
	// Default Connection retry Count
	ConnRetryCount = 20

	// The Plugin Registry singular Instance
	pluginReg *PluginReg = nil
)

/* The Plugin struct that defines the plugin */
type Plugin struct {
	// The URL to reach the Plugin
	PluginUrl string
	// The plugin socket file
	PluginSock string
	// The plugin Connection
	pluginConn *PluginConn.PluginClient
	// The plugin supported function
	methods []string
	// The plugin registered callback
	callbacks map[string]bool
	// Plugin disconnected state (currently Being set but not being used)
	connected bool
	// The Plugin instance PId
	pid int
	// the location of the plugin (It is required while reloading the plugin)
	pluginloc string
}

/* The configuaration for Plugin reg */
type PluginRegConf struct {
	// The location to search for Plugin. Default is .
	PluginLocation string
}

/* PluginReg should be created per types of Plugin
 * each PluginReg monitor a specific location */
type PluginReg struct {
	// The waitgroup to wait for till PluginRegistry doesn't stop
	Wg *sync.WaitGroup
	// The Plugin search location
	PluginLocation string
	// The discovered Plugin location
	discoveredPluginLoc string
	// The mutex to sync the Plugin reg access
	regAccess *sync.Mutex
	// The flag to stop PluginRegistry Service
	stopchan chan int
}

/* Function is called to inititate the PluginRegistry as per the Plugin registry Configuration
   It initiate and return a plugin registry pointer that could be used to manage plugins.
   If Discovery is enabled the DiscoverService Starts */
func PluginRegInit(regConf PluginRegConf) (*PluginReg, error) {

	var wg sync.WaitGroup

	pluginLocation := regConf.PluginLocation

	pluginReg = &PluginReg{}

	// Map to hold discovered Plugins
	pluginReg.DiscoveredPlugin = make(map[string]struct{})

	pluginReg.PluginLocation = pluginLocation
	// Create the discovered plugin location
	// Create discovered plugin location in pluginLocation
	pluginReg.discoveredPluginLoc, direrr = common.CreateDir(pluginLocation, DefaultDiscoveredPlugin)
	if direrr != nil {
		log.ERROR.Printf("Failed to create discovered plugin location, Error : %v", direrr)
		return nil, fmt.Errorf("Failed to create discovered plugin location, Error : %v", direrr)
	}
	pluginReg.Wg = &wg
	pluginReg.RegAccess = &sync.Mutex{}
	pluginReg.stopchan = make(chan int)
	wg.Add(1)
	go pluginReg.discoverPluginService(&wg)
	log.INFO.Printf("Plugin discovery started for : %s", pluginLocation)
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

func (pluginReg *PluginReg) processFile() (isplugin bool, name string, namespace string, version string, untarFold string) {
	var fileName string
	isplugin = false
	if f.IsDir() {
		return
	}
	fileName = f.Name()
	ext := filepath.Ext(fileName)
	// Check if it is a tar File
	if ext != DefaultTarExt {
		return
	}
	// Get the tar name
	tarName := fileName[0 : len(fileName)-len(ext)]
	// Untar the tar file to get the pconf
	tarFile := filepath.Join(pluginLocation, fileName)
	// Untar the file in proper location
	untarErr := common.UntarIt(tarFile, pluginLocation)
	if untarErr != nil {
		log.ERROR.Println("Failed to untar the file: ", tarFile, ", Error: ", untarErr)
		return
	}
	// Get the tar folder
	untarFold = filepath.Join(pluginLocation, tarName)
	// Read the plugin conf
	confFile := filepath.Join(untarFold, DefaultPluginConfFile)
	pluginconf, confloaderror := common.LoadPluginConfigs(confFile)
	if confloaderror != nil {
		log.ERROR.Println("Failed to load plugin Configuration for file: ", tarFile, ", Error: ", confloaderror)
		os.RemoveAll(untarFold)
		return
	}
	if pluginconf.NameSpace != "" && pluginconf.Name != "" && pluginconf.Version != "" {
		isplugin = true
		name = pluginconf.Name
		namespace = pluginconf.NameSpace
		version = plginconf.Version
		return
	}
	return
}

func getKey(name, namespace, version string) string {
	key := fmt.Sprintf("%s_%s_%s", namespace, name, version)
	return key
}

/**
	// Create the plugin id (namespace _ name _ version)
	namespace := pluginconf.NameSpace
	name := pluginconf.Name
	version := pluginconf.Version
	lazyload := pluginconf.LazyLoad
	key := fmt.Sprintf("%s_%s_%s", namespace, name, version)
	// Check if it is discovered
	_, found := pluginreg.DiscoveredMap[key]
	if found {
		log.INFO.Printf("Updating Plugin: %s", key)
	}


	// Create folder named (namespace_name_version) in pluginlocation/discoveredPluginLoc,
	currentpluginlocation, direrr := common.CreateDir(pluginreg.discoveredPluginLoc, key)
	if direrr != nil {
		log.ERROR.Printf("Failed to create plugin location for: %s, Error : %v", key, direrr)
		return nil, fmt.Errorf("Failed to create plugin location for: %s, Error : %v", key, direrr)
	}
	// Copy the tar file content into dir (namespace_name_version)
	copyErr := common.CopyDir(untarFold, currentpluginlocation)
	if copyErr != nil {
		log.ERROR.Printf("Failed to copy plugin files for: %s, Error : %v", key, copyerr)
		return nil, fmt.Errorf("Failed to copy plugin files for: %s, Error : %v", key, direrr)
	}
	// Save the plugin in DiscoveredMap
	pluginreg.DiscoveredMap[key] = currentpluginlocation


**/

/* Function for the routine to discover services */
func (pluginReg *PluginReg) discoverPluginService(wg *sync.WaitGroup) {
	defer wg.Done()
	pluginLocation := pluginReg.PluginLocation

	// Create the filesystem watcher
	watcher, watcheriErr := fsnotify.NewWatcher()
	if watcheriErr != nil {
		log.ERROR.Printf("Failed to initiate watcher, Error : %v", watcheriErr)
		return
	}
	// Start watcher
	watcherDirError := watcher.Watch(pluginLocation)
	if watcherDirError != nil {
		log.ERROR.Printf("Failed to start watch on %s: , Error : %v", pluginLocation, watcherDirError)
		return
	}
	// For holding the last created file in the dir and create time
	createtime := 0
	createfile := ""
	// For holding the last modified file and the modify time
	modifytime := 0
	modifyfile := ""
	for {
		select {
		case event := watcher.Event:
			switch {
			case ev.IsCreate():
				createtime = time.Now().Second()
				createfile = ev.Name
				fmt.Println("Create Event: ", ev.Name)
				isplugin, name, namespace, version, untarfold := pluginReg.processFile(createfile)
				if !isplugin {
					continue
				}
				key := getKey(name, namespace, version)
				if Is
				if !pluginreg.lazyload {
					pluginreg.LoadPlugin(namespace, name, version)
				}
			case ev.IsModify():
				currenttime := time.Now().Second()
				if (ev.Name == createfile && (currenttime-createtime) < 1) || (ev.Name == modifyfile) && (currenttime-modifytime) < 1 {
				} else {
					modifytime = time.Now().Nanosecond()
					modifyfile = ev.Name
					fmt.Println("Modify Event: ", ev.Name)
					pluginReg.processFile(modifyfile)
					// TODO: Check if plugin was already loaded or lazy load is not configured
					if !pluginreg.lazyload /*&& pluginreg.IsLoaded()  */ {
						pluginreg.LoadPlugin(namespace, name, version)
					}

				}
			case ev.IsDelete():
				fmt.Println("Delete Event: ", ev.Name)
				if ev.Name == pluginLocation {
					log.ERROR.Printf("Plugin Location has been removed: %s", pluginLocation)
					return
				}
			}
		case watchererr := <-watcher.Error:
			log.ERROR.Printf("Error while watching on %s: , Error : %v", pluginLocation, watchererr)
			return
		case pluginReg.stopchan:
			log.Info.Printf("Stopping PluginReg Channel")
			return
		}
	}
}

/* Check if a plugin is discovered by the plugin registry discovery service automatically or is discover implicitly */
func (pluginReg *PluginReg) IsDiscovered(pluginname string) bool {

	return pluginReg.isDiscovered(pluginname)
}

/* Internal: Check if a plugin is already discovered */
func (pluginReg *PluginReg) isDiscovered(appPlugin string) bool {
	_, pluginDiscovered := pluginReg.DiscoveredPlugin[appPlugin]
	if !pluginDiscovered {
		return false
	}
	return true
}

/* Unload a Plugin from the plugin Registry. It invokes a stop request to the plugin.
   (It doesn't remove the Plugin from Discovered Plugin List) */
func (plugin *Plugin) UnloadPlugin() error {

	// Initiate Locking
	//pluginReg.RegAccess.Lock()
	//defer pluginReg.RegAccess.Unlock()

	// Send the Stop request
	stopErr := plugin.stop()
	if stopErr != nil {
		log.ERROR.Println("Failed to send stop to the plugin: ", stopErr)
	}

	// Close the connection
	plugin.pluginConn.Close()

	// Kill the plugin process
	stoppErr := stopProcess(plugin.pid)
	if stoppErr != nil {
		log.ERROR.Println("Failed to stop the plugin process: ", stoppErr)
	}

	return nil
}

/* Function to reload a plugin */
func (plugin *Plugin) ReloadPlugin() error {

	plugin.UnloadPlugin()

	newPlugin, err := pluginReg.LoadPluginInstance(plugin.pluginloc)
	if err != nil {
		return fmt.Errorf("Failed to reload plugin: %v", err)
	}

	// The plugin Connection
	plugin.pluginConn = newPlugin.pluginConn
	// Plugin disconnected state (currently Being set but not being used)
	plugin.connected = newPlugin.connected
	// The Plugin instance PId
	plugin.pid = newPlugin.pid

	return nil
}

func stopProcess(pid int) error {

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("Failed to get Process of Id: %d", pid)
	}

	// Kill the procecss
	killErr := process.Signal(syscall.SIGUSR1)
	if killErr != nil {
		return fmt.Errorf("Failed to deliver SIGUSR1 to process %d: %v", pid, killErr)
	}

	return nil
}

/* Load the plugin to the plugin Registry explicitly when lazy load is active.
(if The discovery Process is not running, It search for the plugin and then load it to the registry)
*/
func (pluginReg *PluginReg) LoadPlugin(namespace string, name string, version string) (*Plugin, error) {

	// Get the plugin tar location
	tarFold := pluginLoc

	// Runtime Conf file
	confFile := filepath.Join(tarFold, DefaultPluginiRuntimeConfFile)

	// Create RuntimeConf
	pluginConf := common.RuntimeConf{}
	StartPath := "./" + PluginBinary
	pluginConf.Url = PluginUrl
	pluginConf.Sock = PluginSockFile

	// Save new plugin Conf
	confSaveError := common.SaveRuntimeConfigs(confFile, pluginConf)
	if confSaveError != nil {
		log.ERROR.Println("Configuration load failed for file: ", confFile, ", Error: ", confSaveError)
		return nil, SaveConfError
	}

	// get the start path
	startPath := filepath.Join(tarFold, StartPath)

	// Start the Plugin
	log.DEBUG.Printf("Starting plugin: %s\n", startPath)
	pid, startErr := pluginReg.startPlugin(startPath)
	if startErr != nil {
		log.ERROR.Println("Failed to start the plugin: ", startErr)
	}

	// get the unix socket file path
	sockFile := filepath.Join(tarFold, pluginConf.Sock)

	retryCount := 0
	var pluginConn *PluginConn.PluginClient = nil
	time.Sleep(DefaultInterval * 4)
	for retryCount < ConnRetryCount {
		var connErr error
		// Initiate Connection to a Plugin
		log.DEBUG.Printf("Trying to connect: %s\n", sockFile)
		pluginConn, connErr = PluginConn.NewPluginClient(sockFile)
		if connErr == nil {
			break
		}
		retryCount++
		// Sleep for a delay
		time.Sleep(DefaultInterval)
	}
	if pluginConn == nil {
		return nil, PluginConnFailed
	}

	plugin := &Plugin{}
	plugin.PluginSock = sockFile
	plugin.PluginUrl = pluginConf.Url
	plugin.pluginConn = pluginConn
	plugin.connected = true
	plugin.callbacks = make(map[string]bool)
	// set the plugin instance process id
	plugin.pid = pid
	plugin.pluginloc = pluginLoc

	// Activate the plugin
	activateErr := plugin.activate()
	if activateErr != nil {
		return plugin, activateErr
	}

	return plugin, nil
}

func (pluginReg *PluginReg) startPlugin(startFile string) (int, error) {

	// Change the file permission
	err := os.Chmod(startFile, 0777)
	if err != nil {
		log.DEBUG.Printf("Failed to change mode: %v", err)
		return 0, err
	}

	dir := filepath.Dir(startFile)
	startPath, _ := filepath.Abs(dir)
	file := path.Base(startFile)

	_, lookErr := exec.LookPath(startFile)
	if lookErr != nil {
		log.DEBUG.Printf("Lookerror")
		return 0, lookErr
	}
	env := os.Environ()
	attr := &syscall.ProcAttr{Dir: startPath, Env: env}
	pid, execErr := syscall.ForkExec(file, nil, attr)
	if execErr != nil {
		log.DEBUG.Printf("Exeerror")
		return 0, execErr
	}
	log.DEBUG.Printf("Started process: %d\n", pid)
	return pid, nil
}

// function to check plugin status
func (plugin *Plugin) checkConnection() bool {
	if plugin.Ping() != nil {
		return false
	}
	return true
}

func (plugin *Plugin) ReConnect() error {

	// Connect to the plugin
	pluginConn, connErr := PluginConn.NewPluginClient(plugin.PluginSock)
	if connErr != nil {
		plugin.connected = false
		return fmt.Errorf("Failed to reconnect: %v", connErr)
	}
	// Set connection object
	plugin.pluginConn = pluginConn
	plugin.connected = true

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

	if !plugin.connected {
		return fmt.Errorf("Plugin is not connected")
	}

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
func (plugin *Plugin) Execute(funcName string, args ...interface{}) (error, []interface{}) {

	if !plugin.connected {
		return fmt.Errorf("Plugin is not connected"), nil
	}

	found := false
	// check if method is registered
	for _, method := range plugin.methods {
		if method == funcName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("Method of name : %s is not registered", funcName), nil
	}

	pluginUrl := plugin.PluginUrl
	pluginConn := plugin.pluginConn

	requestUrl := pluginUrl + "/" + funcName
	data := CreateJson(args)
	request := &PluginConn.PluginRequest{Url: requestUrl, Body: data}

	resp, err := pluginConn.Request(request)
	if err != nil {
		plugin.connected = false
		// try to reconnect the plugin
		err := plugin.ReConnect()
		if err != nil {
			err = plugin.ReloadPlugin()
		}
		if err != nil {
			return fmt.Errorf("Failed to communicate with plugin"), nil
		}
	}
	if resp.Status != "200 OK" {
		return fmt.Errorf("request failed"), nil
	}

	ret := resp.Body

	if string(resp.Body) == "<nil>" {
		ret = nil
	}

	return nil, ret
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
