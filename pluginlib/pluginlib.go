package pluginlib

import (
	"fmt"
	"github.com/swarvanusg/GoPlug"
	"os"
	"os/signal"
	"syscall"
)

// The alleydog plugin Impl
type alleydogPluginImpl struct {
	pluginReg *pluginmanager.PluginImpl
	// :ifecycle Plugin info
	lifecycleInstanceRegisterer func(*LifecycleInitConf) (LifecyclePluginImpl, error)
	lifecycleInstanceMap        map[string]interface{}
	// TODO: More plugin type should follow here
}

// The lifecycle plugin Impl -- Plugin writer should implement the interface
type LifecyclePluginImpl interface {
	Start() (pid_cid string, err error)
	Stop() error
}

// The Lifecycle Plugin Conf that will be passsed to the instance Registerer
type LifecycleInitConf struct {
	CIL      string
	Deploy   string
	PortMap  map[string]string
	InitData []byte
}

// Const that is used for deploy type
const (
	DEPLOY_PROCESS   = "process"
	DEPLOY_CONTAINER = "container"
)

// TODO: More plugin type should follow here

type PluginConfig struct {
	// Method that creates a new instance of LifecyclePlugin for as per the Conf
	LifecycleInstanceRegisterer func(*LifecycleInitConf) (LifecyclePluginImpl, error)
	// TODO: More plugin type should follow here
}

var (
	alleydogPlugin *alleydogPluginImpl = nil
)

// Function to register a plugin
func RegisterPlugin(conf PluginConfig) (*alleydogPluginImpl, error) {

	pluginConf := pluginmanager.PluginImplConf{PluginLoc: pluginmanager.DefaultPluginConfFile, Activator: pluginStarter, Stopper: pluginStopper}
	// Implement the Plugin
	regPlugin, pluginInitError := pluginmanager.PluginInit(pluginConf)
	if pluginInitError != nil {
		return nil, fmt.Errorf("Failed to initialize the plugin: %v", pluginInitError)
	}

	// To allow the same object to be used repetedly
	if alleydogPlugin == nil {
		alleydogPlugin = &alleydogPluginImpl{}
	}

	// Check if the lifecycle implementation provided
	if conf.LifecycleInstanceRegisterer != nil {
		alleydogPlugin.lifecycleInstanceRegisterer = conf.LifecycleInstanceRegisterer
		alleydogPlugin.lifecycleInstanceMap = make(map[string]interface{})

		// Register alleydog lifecycle method
		regPlugin.RegisterMethod(lifecycleInit)
		regPlugin.RegisterMethod(lifecycleStart)
		regPlugin.RegisterMethod(lifecycleStop)
	}
	// TODO: More plugin type should follow here

	alleydogPlugin.pluginReg = regPlugin

	return alleydogPlugin, nil
}

func pluginStarter(data []byte) []byte {
	return nil
}

func pluginStopper(data []byte) []byte {
	return nil
}

func lifecycleInit(reqdata []byte) (response []byte) {

	cid, cil, deploy, portmap, data, decodeErr := pluginmanager.DecapsuleInitRequest(reqdata)
	if decodeErr != nil {
		response, _ = pluginmanager.EncapsulePluginResponse(cid, false, "", fmt.Sprintf("Failed to decapsule init param: %s", decodeErr))
		return
	}

	// The the lifecycle Configuration
	lifecycleConf := LifecycleInitConf{cil, deploy, portmap, data}
	var lifecycleinstance LifecyclePluginImpl = nil

	// Check if the plugin is registered as a lifecycle plugin
	if alleydogPlugin.lifecycleInstanceRegisterer != nil {
		var initerr error
		lifecycleinstance, initerr = alleydogPlugin.lifecycleInstanceRegisterer(&lifecycleConf)
		if initerr != nil {
			response, _ = pluginmanager.EncapsulePluginResponse(cid, false, "", fmt.Sprintf("Failed to initialize the Application instance: %s", initerr))
			return
		}
	} else {
		response, _ = pluginmanager.EncapsulePluginResponse(cid, true, "", fmt.Sprintf("The plugin is not restered any Lifecycle implimentation: contact to the plugin admin"))
		return
	}

	alleydogPlugin.lifecycleInstanceMap[cid] = lifecycleinstance

	response, _ = pluginmanager.EncapsulePluginResponse(cid, true, "", "")
	return
}

func lifecycleStart(reqdata []byte) (response []byte) {

	cid, decodeerr := pluginmanager.DecapsuleManageRequest(reqdata)
	if decodeerr != nil {
		response, _ = pluginmanager.EncapsulePluginResponse(cid, false, "", fmt.Sprintf("Failed to decapsule start request: %s", decodeerr))
		return
	}

	// Get the lifecycleinstance from the map
	applicationInstance, found := alleydogPlugin.lifecycleInstanceMap[cid]
	if !found {
		response, _ = pluginmanager.EncapsulePluginResponse(cid, false, "", fmt.Sprintf(pluginmanager.NoInstance))
		return
	}

	lifecycleInstance := applicationInstance.(LifecyclePluginImpl)

	pid_cid, err := lifecycleInstance.Start()
	if err != nil {
		response, _ = pluginmanager.EncapsulePluginResponse(cid, false, pid_cid, fmt.Sprintf("%v", err))
		return
	}

	response, _ = pluginmanager.EncapsulePluginResponse(cid, true, pid_cid, "")

	return
}

func lifecycleStop(reqdata []byte) (response []byte) {

	cid, decodeerr := pluginmanager.DecapsuleManageRequest(reqdata)
	if decodeerr != nil {
		response, _ = pluginmanager.EncapsulePluginResponse(cid, false, "", fmt.Sprintf("Failed to decapsule stop request: %s", decodeerr))
		return
	}

	// Get the lifecycleinstance from the map
	applicationInstance, found := alleydogPlugin.lifecycleInstanceMap[cid]
	if !found {
		response, _ = pluginmanager.EncapsulePluginResponse(cid, false, "", fmt.Sprintf(pluginmanager.NoInstance))
		return
	}

	lifecycleInstance := applicationInstance.(LifecyclePluginImpl)

	err := lifecycleInstance.Stop()

	if err != nil {
		response, _ = pluginmanager.EncapsulePluginResponse(cid, false, "", fmt.Sprintf("%v", err))
	}

	response, _ = pluginmanager.EncapsulePluginResponse(cid, true, "", "")

	return
}

// Function to start a plugin
func (plugin *alleydogPluginImpl) StartPlugin() error {

	plugReg := plugin.pluginReg

	// Start the plugin
	(*plugReg).Start()
	return nil
}

// Function to wait for a plugin to stop. The wait finish when the plugin gets a SIGUSR1 from agent or a explicit SIGTERM
func (plugin *alleydogPluginImpl) WaitForPluginStop() error {
	pluginExitChannel := makeExitChannel()
	// We block on this channel
	<-pluginExitChannel
	plugReg := plugin.pluginReg
	// Stop the plugin server
	(*plugReg).Stop()
	return nil
}

func makeExitChannel() chan os.Signal {
	//channel for catching signals of interest
	signalCatchingChannel := make(chan os.Signal)

	//catch Ctrl-C and Kill -30 <pid> signals
	signal.Notify(signalCatchingChannel, syscall.SIGUSR1, syscall.SIGTERM)

	return signalCatchingChannel
}
