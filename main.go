package main

import (
	"bufio"
	"com.ss/goplugin/PluginReg"
	"fmt"
	"os"
)

func main() {
	plugRegConf := PluginReg.PluginRegConf{PluginLocation: "Plugin", AutoDiscover: true, ConfExt: ".conf"}
	pluginReg, regErr := PluginReg.PluginRegInit(plugRegConf)
	if regErr != nil {
		fmt.Printf("Plugin reg init failed")
		return
	}

	/* Wait for input */
	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	/* Check if plug1 is loaded */
	if pluginReg.IsLoaded("Plug1", "Monitor") {
		fmt.Printf("plug1 is loaded")
		return
	}

	/* Load plug1, as lazy load is configured */
	plug1, err := pluginReg.LoadPlugin("Plug1", "Monitor")
	if err != nil {
		fmt.Printf("Plugin loading failed : %s\n", err)
		return

	}
	fmt.Printf("Loaded Plugin: %v\n", plug1)

	/* get a plugin */
	plugin1 := pluginReg.GetPlugin("Plug1", "Monitor")
	if plugin1 == nil {
		fmt.Printf("Get a plugin failed")
		return
	}

	/* Wait for input */
	fmt.Print("Press 'Enter' to continue...")
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
