package main

import (
	"bufio"
	"fmt"
	GoPlug "github.com/swarvanusg/GoPlug"
	"os"
)

func main() {
	plugRegConf := GoPlug.PluginRegConf{PluginLocation: "Plugin"}
	pluginReg, regErr := GoPlug.PluginRegInit(plugRegConf)
	if regErr != nil {
		fmt.Printf("Plugin reg init failed\n")
		return
	}

	fmt.Println("Press 'Enter' to continue...\n")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	/* get a plugin */
	plugin1 := pluginReg.GetPlugin("Do", "Test")
	if plugin1 == nil {
		fmt.Printf("Get a plugin failed\n")
		return
	}

	/* Check if plugin is connected */
	pingErr := plugin1.Ping()
	if pingErr != nil {
		fmt.Printf("Ping Failed - Plugin is not connected: %v\n", pingErr)
		return
	}
	fmt.Printf("Plugin is connected")

	methods := plugin1.GetMethods()
	/* get the list of method */
	fmt.Println("Available Methods: ", methods, "\n")
	for _, value := range methods {
		plugin1.Execute(value, nil)
	}

	/* Register a callback */
	err := plugin1.RegisterCallback(callBack)
	if err != nil {
		fmt.Println("Call back registration failed: ", err)
		return
	}

	/* Wait for input */
	fmt.Printf("Press 'Enter' to unload plugin\n")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	/* unload the Plugin */
	unloadErr := pluginReg.UnloadPlugin(plugin1)
	if unloadErr != nil {
		fmt.Printf("Unload of a plugin failed")
		return
	}

	pluginReg.Stop()
	fmt.Printf("Waiting for pluginReg to stop \n")
	pluginReg.WaitForStop()
}

func callBack(data []byte) {
	fmt.Println("Executing callback: ", string(data))
}
