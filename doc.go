/*


< GoPlug >
 --------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
                
GoPlug is a pure Go Plugin libary project that provides flexibility,
Loose Coupling and moduler approach of Building Software in/around Go. 
The goal of the project is to provide a simple, fast and a reliable 
plugin architecture that is independent of the platform.


Lifecycle

GoPlug plugin lifecycle is quite simple as it consist only three state.  
1. Stopped : Plugin is not yet started or stopped  
2. Discovered/Installed : Plugin is discoved and ready to be started  
3. Started/Loaded : Plugin is started or Loaded for serving request  


Plugin Registry

Each of the application creates a Plugin Registry to manage Plugins. 
Plugin Registry is based on plugin discovery service that provide api 
to search, load and unload plugin to/from registry.
Auto discovery service at plugin registry could be disabled resulting 
plugin to be discovered at loading time.


Plugin

Each plugin makes itself available for the discovery service, and while 
discovered it is loaded by the application. On a successful loading 
start() is called and on a successful uploading stop() is called

Lazy start could be enabled to make plugin loaded by explicit call to 
Plugin Registry rather than at discovery.


How It Works

Plugins runs a different process that sould be started explicitly. Unix 
domain socket is used for IPC where the communication is based on HTTP 
request response model.

Step by Step:
1> At start of the Plugin it opens a Unix domain socket and listen for connection  
2> Once it initialized it puts the .pconf file in a specific location of Plugin Discovery  
3> Plugin Registry discover the .pconf and load the configuration to get the properties and UNIX sock  
4> Plugin Registry initialize the Connections using UNIX sock and it loads the Plugin information  
5> Http request is made as per the methods Executed over the connection  

*/
package GoPlug
