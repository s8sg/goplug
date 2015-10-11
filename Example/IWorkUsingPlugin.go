package main

import (
	"bufio"
	"fmt"
	GoPlug "github.com/swarvanusg/GoPlug"
	"os"
)

func main() {
	plugRegConf := GoPlug.PluginRegConf{PluginLocation: "Plugin", AutoDiscover: true, ConfExt: ".pconf"}
	pluginReg, regErr := GoPlug.PluginRegInit(plugRegConf)
	if regErr != nil {
		fmt.Printf("Plugin reg init failed")
		return
	}

	/* Wait for input */
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	/* Check if plug1 is loaded */
	/*
		if pluginReg.IsLoaded("Do", "Test") {
			fmt.Printf("plug1 is loaded")
			return
		}*/

	/* Load plug1, as lazy load is configured */
	/*
		plug1, err := pluginReg.LoadPlugin("Plug1", "Monitor")
		if err != nil {
			fmt.Printf("Plugin loading failed : %s\n", err)
			return

		}*/

	/* get a plugin */
	plugin1 := pluginReg.GetPlugin("Do", "Test")
	if plugin1 == nil {
		fmt.Printf("Get a plugin failed")
		return
	}

	methods := plugin1.GetMethods()
	/* get the list of method */
	fmt.Println("Methods: ", methods, "\n")
	/*
		for _, value := range methods {
			plugin1.Execute(value, nil)
		}
	*/

	/* Wait for input */
	fmt.Print("Press 'Enter' to Register Callback...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	/* Register a callback */
	err := plugin1.RegisterCallback(callBack)
	if err != nil {
		fmt.Println("Call back registration failed: ", err)
		return
	}

	/* Wait for input */
	fmt.Print("Press 'Enter' to unload plugin")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	/* unload the Plugin */
	unloadErr := pluginReg.UnloadPlugin(plugin1)
	if unloadErr != nil {
		fmt.Printf("Unload of a plugin failed")
		return
	}

	pluginReg.Stop()
	pluginReg.WaitForStop()
}

func callBack(data []byte) {
	fmt.Println("Executing callback: ", string(data))
}
