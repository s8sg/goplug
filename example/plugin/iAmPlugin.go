package main

import (
	"bufio"
	"fmt"
	GoPlug "github.com/swarvanusg/GoPlug/pluginlib"
	"os"
	"os/signal"
	"syscall"
)

var pluginimpl *GoPlug.PluginImpl
var stopchan chan (int)

type DoPlugin struct {
	// My dummy plugin
}

func (plugin DoPlugin) Init() {
	fmt.Printf("Plugin has been initialized")
}

func (plugin DoPlugin) Start(data []byte) []byte {
	fmt.Printf("Starting Plugin\n")
	return nil
}

func (plugin DoPlugin) Stop(data []byte) []byte {
	fmt.Printf("Stoping Plugin\n")
	stopchan <- 1
	return nil
}

func Do(data []byte) []byte {
	fmt.Printf("I'm Doing\n")
	return nil
}

func main() {

	var err error
	var plugin DoPlugin
	stopchan = make(chan int)
	pluginimpl, err = GoPlug.PluginInit(plugin)
	if err != nil {
		fmt.Printf("Plugin Init Error: %s\n", err)
		return
	}
	// Register Do Method
	pluginimpl.RegisterMethod(Do)
	// Start the plugin
	pluginimpl.Start()
	shutdownChannel := makeShutdownChannel()

	/* Wait for input */
	fmt.Println("Press 'Enter' to notify registered Callback")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	//notify the callback
	pluginimpl.Notify("main.callBack", []byte("Test Data"))

	//we block on this channel
	<-stopchan
	// Stop the plugin
	pluginimpl.Stop()
}
