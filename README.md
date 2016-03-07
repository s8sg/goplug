# GoPlug 

![](https://github.com/swarvanusg/goplug/blob/master/doc/goplug_v2.png)

GoPlug is a pure Go Plugin library project that provides **flexibility**, **loose Coupling** and **moduler approach** of Building Software in/around Go. The goal of the project is to provide a simple, fast and a reliable plugin architecture that is independent of the platform. 

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg?style=flat-square)](https://godoc.org/github.com/swarvanusg/goplug)

### Version
0.1.0

### Usage

#### Step 1 : Get It
To get the GoPlug install Go and execute the below command 
```
go get github.com/swarvanusg/GoPlug
```

#### Step 2: Lifecycle
GoPlug plugin life-cycle is quite simple as it consist only three state. 
1.  **Stopped** : Plugin is not yet started or stopped
2.  **Discovered/Installed** : Plugin is discover and ready to be started
3.  **Started/Loaded** : Plugin is started or Loaded for serving request

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
```json
    {
        "Name" : "NameOfPlugin",
        "NameSpace" : "NamespaceOfPlugin",
        "Url" : "unix://PluginUrl",
        "sock" : "unixSockLocation.sock",
        "LazyLoad" : false,
    }
```
##### Application That Use Plugins
___
![](https://github.com/swarvanusg/goplug/blob/master/doc/goplug_app.png)

Plugin registry is initialized with the plugin location where it will search for plugin conf **(.pconf)**, along with the Auto Discover setting. If auto discovery is enabled the discover service starts and search for new plugin, while in other case of discovery service not running, plugin gets discovered while loading (via Explicit call to LoadPlugin) if available.
```go
    plugRegConf := GoPlug.PluginRegConf{PluginLocation: "./PluginLoc", AutoDiscover: true}
    /* Initialize a Plugin Registry that will search location "./PluginLoc" for '.pconf' file */  
    pluginReg, err := GoPlug.PluginRegInit(plugRegConf)
```
Lazyload is a feature that prevents auto loading of a plugin when it is discovered. If Plugin is Configured for lazy load plugin should be loaded explicitly when needed by the user.  

```go
    plugin, err := pluginReg.LoadPlugin("name", "namespace")
```
Each plugin is identified by the plugin name and namespace
```go
    plugin := pluginReg.GetPlugin("name", "namespace")
```
Plugin can be searched for available methods (registered methods by Plugin implementation)
```go
    methodList := plugin.GetMethods()
```
Method could be executed by method name 
```go
    returnBytes, err := plugin.Execute(methodName, inputBytes)
```
Callback could be registered in Apllication to receive notification from plugin
```go
    plugin.RegisterCallback(Foo)
    ...
    func Foo(data []byte) {
        // Callback body called on notification from pugin
    }
```
Plugin could be forced to unload or stopped
```go
    err := pluginReg.UnloadPlugin(plugin)
```
##### Plugin Implementation
___
![](https://github.com/swarvanusg/goplug/blob/master/doc/goplug_plugin.png)

Plugin is initialized with the **Location**, **Name**, **Namespace** (optional), **Url** (optional), **LazyStart conf**, **Activator** and **Stopper**. 
The Plugin location should be same on which Plugin Registry is configured
```go
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
```go
plugin.RegisterMethod(Do)
...
func Do(input []byte) []byte {
    // Call on execution of "Do" from application
}
```
Plugin start makes the plugin available for the discovery service and to be loaded
```go
plugin.Start()
```
Plugin could notify application using callback. A list of registered callbacks are available at plugins
```go
    //get available callback list
    callbackList := plugin.GetCallbacks()
    ...
    err := plugin.Notify(callbackName, inputBytes)
```
Plugin stop makes the plugin to be stopped and unavailable from the Plugin Reg service. It should be done after plugin is unloaded from the Plugin registry. 
```go
plugin.stop()
```
[More ...](https://godoc.org/github.com/swarvanusg/GoPlug#pkg-index)

#### Step 4: How It Works
Plugins runs as a different process that is started by the plugin registry. For IPC in Linux Unix domain socket is used, where in Windows com is used. The communication is based on HTTP request response model. 

### Current Status
GoPlug is unstable and in active development and testing

### Future Scope
As GoPlug Plugin are independent process and the communication is based on Unix socket and HTTP. Plugin could be developed using any programming language. In future GoPlug Plugin Implementation library should be implemented in different languages. Which will allow plugins to be written in different languages.  

### More Information
This is an early release. I’ve been using it for a while and this is working fine. I like this one pretty well, but no guarantees that it won’t change a bit. 
