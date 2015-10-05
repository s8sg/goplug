# GoPlug

GoPlug is a pure Go Plugin libary project that provides flexibility, Loose Coupling and moduler approach of Building Software in/around Go. The goal of the project is to provide a simple, fast and a reliable plugin architecture that is independent of the platform. 

### Version
0.1.0

### Usage

#### Step 1 : Get It
```
go get github.com/swarvanusg/GoPlug
```

#### Step 2: Lifecycle
GoPlug plugin lifecycle is quite simple as it consist only three state. 
1. **Stopped** : Plugin is not yet started or stopped
2. **Discovered/Installed** : Plugin is discoved and ready to be started
3. **Started/Loaded** : Plugin is started or Loaded for serving request

Each of the application creates a Plugin Registry to manage Plugins. Plugin Registry is based on plugin discovery service that provide api to search, load and unload plugin to/from registry.

Auto discovery service at plugin registry could be disabled resulting plugin to be discovered at loading time.

Each plugin makes itself available for the discovery service, and while discovered it is loaded by the application. On a successful loading start() is called and on a successful uploading stop() is called

Lazy start could be enabled to make plugin loaded by explicit call rather than at discovery. 


#### Step 3: Use it  

##### Application That Use Plugins
___
Plugin registry is initialized with the plugin location where it will search for plugin conf **(.pconf)**, along with the Auto Discover settings.   
```
    plugRegConf := GoPlug.PluginRegConf{PluginLocation: "./PluginLoc", AutoDiscover: true}
    /* Initialize a Plugin Registry that will search location "./PluginLoc" for '.pconf' file */  
    pluginReg, err := GoPlug.PluginRegInit(plugRegConf)
```
If Plugin is Configured for lazy load plugin should be loaded explicitly. In case of discovery service not running, plugin gets discovered before loading if available. 

```
    plugin, err := pluginReg.LoadPlugin("name", "namespace")
```
Each plugin is identified by the plugin name and namespace
```
    plugin := pluginReg.GetPlugin("name", "namespace")
```
Plugin can be searched for available method (registered method by plugin implementation)
```
    methodList := plugin.GetMethods()
```
Method could be executed by method name 
```
    returnBytes, err := plugin.Execute(method, inputBytes)
```
Plugin could be forced to unload or stopped
```
    err := pluginReg.UnloadPlugin(plugin)
```
##### Plugin Implementation
___
Plugin is initialized with the **Location**, **Name**, **Namespace** (optional), **Url** (optional), **LazyStart conf**, **Activator** and **Stopper**. 
The Plugin location should be same on which Plugin Registry is configured
```
    config := GoPlug.PluginImplConf{"PluginLoc", "Name", "Namespace", "unix://URL", false, activate, stop}
    plugin, err := GoPlug.PluginInit(config)
    ...
    func activate(input []byte) []byte {
        // Called on Activation of the Plugin
    }
    func stop(input []byte) []byte {
        // Called on Deactivation of the Plugin
    }
```
Method should be registered before starting the plugin
```
plugin.RegisterMethod("Do", Do)
...
func Do(input []byte) []byte {
    // Call on execution of "Do" 
}
```
Plugin start makes the plugin available for the discovery service and to be loaded
```
plugin.Start()
```
Plugin stop makes the plugin to be stopped and unavailable from the Plugin Reg service. It should be done after plugin is unloaded from the Plugin registry. 
```
plugin.stop()
```
#### Step 3: How It Works
Plugins runs a different process that sould be started explicitly. Unix domain socket is used for IPC where the communication is based on HTTP request response model. 
###### Step by Step:
1. Plugin start and open a Unix domain socket and listen
2. It puts the .pconf file in the location for discovery
3. Plugin Registry discover the .pconf and load the configuration to get the UNIX socket file
4. Connection using UNIX socket is created by the Plugin Registry
5. Http request is made as per the methods over the connection

### More Information
This is an early release. I’ve been using it for a while and this is working fine. I like this one pretty well, but no guarantees
that it won’t change a bit. 

For Any more info Contact:
```
swarvanusg@gmail.com
```
