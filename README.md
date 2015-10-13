# GoPlug ![Build Status](https://travis-ci.org/go-sql-driver/mysql.png?branch=master)
```
< GoPlug >
 --------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
```

GoPlug is a pure Go Plugin libary project that provides flexibility, Loose Coupling and moduler approach of Building Software in/around Go. The goal of the project is to provide a simple, fast and a reliable plugin architecture that is independent of the platform. 

[GpPlug GoDoc] (https://godoc.org/github.com/swarvanusg/GoPlug)

### Version
0.1.0

### Usage

#### Step 1 : Get It
To get the GoPlug install Go and execute the below command 
```
go get github.com/swarvanusg/GoPlug
```

#### Step 2: Lifecycle
GoPlug plugin lifecycle is quite simple as it consist only three state. 
1. **Stopped** : Plugin is not yet started or stopped
2. **Discovered/Installed** : Plugin is discoved and ready to be started
3. **Started/Loaded** : Plugin is started or Loaded for serving request

###### Plugin Registry
Each of the application creates a Plugin Registry to manage Plugins. Plugin Registry is based on plugin discovery service that provide api to search, load and unload plugin to/from registry.

Auto discovery service at plugin registry could be disabled resulting plugin to be discovered at loading time.

###### Plugin
Each plugin makes itself available for the discovery service, and while discovered it is loaded by the application. On a successful loading start() is called and on a successful uploading stop() is called

Lazy start could be enabled to make plugin loaded by explicit call to Plugin Registry rather than at discovery. 

#### Step 3: Use it  
##### Plugin Conf
___
Plugin conf (.pconf) defines the plugin properties. It is created by the Plugins at Plugin startup and loaded by the Application. 
###### Example.pconf
```
    {
        "Name" : NameOfPlugin
        "NameSpace" : NamespaceOfPlugin
        "Url" : unix://PluginUrl
        "sock" : unixSockLocation.sock
        "LazyLoad" : false
    }
```
##### Application That Use Plugins
___
Plugin registry is initialized with the plugin location where it will search for plugin conf **(.pconf)**, along with the Auto Discover setting. If auto discovery is enabled the discover service starts and search for new plugin, while in other case of discovery service not running, plugin gets discovered while loading (via Explicit call to LoadPlugin) if available.
```
    plugRegConf := GoPlug.PluginRegConf{PluginLocation: "./PluginLoc", AutoDiscover: true}
    /* Initialize a Plugin Registry that will search location "./PluginLoc" for '.pconf' file */  
    pluginReg, err := GoPlug.PluginRegInit(plugRegConf)
```
Lazyload is a feature that prevents auto loading of a plugin when it is discovered. If Plugin is Configured for lazy load plugin should be loaded explicitly when needed by the user.  

```
    plugin, err := pluginReg.LoadPlugin("name", "namespace")
```
Each plugin is identified by the plugin name and namespace
```
    plugin := pluginReg.GetPlugin("name", "namespace")
```
Plugin can be searched for available methods (registered methods by Plugin implementation)
```
    methodList := plugin.GetMethods()
```
Method could be executed by method name 
```
    returnBytes, err := plugin.Execute(methodName, inputBytes)
```
Callback could be registered in Apllication to receive notification from plugin
```
    plugin.RegisterCallback(Foo)
    ...
    func Foo(data []byte) {
        // Callback body called on notification from pugin
    }
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
plugin.RegisterMethod(Do)
...
func Do(input []byte) []byte {
    // Call on execution of "Do" from application
}
```
Plugin start makes the plugin available for the discovery service and to be loaded
```
plugin.Start()
```
Plugin could notify application using callback. A list of registered callbacks are available at plugins
```
    //get available callback list
    callbackList := plugin.GetCallbacks()
    ...
    err := plugin.Notify(callbackName, inputBytes)
```
Plugin stop makes the plugin to be stopped and unavailable from the Plugin Reg service. It should be done after plugin is unloaded from the Plugin registry. 
```
plugin.stop()
```
[More ...](https://godoc.org/github.com/swarvanusg/GoPlug#pkg-index)

#### Step 4: How It Works
Plugins runs a different process that sould be started explicitly. Unix domain socket is used for IPC where the communication is based on HTTP request response model. 
###### Step by Step:
1. At start of the Plugin it opens a Unix domain socket and listen for connection
2. Once it initialized it puts the .pconf file in a specific location of Plugin Discovery
3. Plugin Registry discover the .pconf and load the configuration to get the properties and UNIX sock
4. Plugin Registry initialize the Connections using UNIX sock and it loads the Plugin information 
5. Http request is made as per the methods Executed over the connection

### Current Status
GoPlug is unstable and in active development and testing

### Future Scope
As GoPlug Plugin are independent process and the communication is based on Unix socket and Http. Plugin could be developed using any programming language. In future GoPlug Plugin Implementation library should be implementated in different languages.  

### More Information
This is an early release. I’ve been using it for a while and this is working fine. I like this one pretty well, but no guarantees
that it won’t change a bit. 

For Any more info Contact:
```
swarvanusg@gmail.com
```
