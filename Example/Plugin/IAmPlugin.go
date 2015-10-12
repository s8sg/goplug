package main

import (
	"bufio"
	"fmt"
	GoPlug "github.com/swarvanusg/GoPlug"
	"os"
	"os/signal"
	"syscall"
)

var plugin *GoPlug.PluginImpl

func main() {

	var err error
	config := GoPlug.PluginImplConf{"./", "Do", "Test", "", false, activate, stop}
	plugin, err = GoPlug.PluginInit(config)
	if err != nil {
		fmt.Printf("Plugin Init Error: %s\n", err)
		return
	}
	plugin.RegisterMethod(Do)
	plugin.Start()
	shutdownChannel := makeShutdownChannel()

	/* Wait for input */
	fmt.Println("Press 'Enter' to notify registered Callback")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	//notify the callback
	plugin.Notify("main.callBack", []byte("Test Data"))

	//we block on this channel
	<-shutdownChannel
	plugin.Stop()
}

func makeShutdownChannel() chan os.Signal {
	//channel for catching signals of interest
	signalCatchingChannel := make(chan os.Signal)

	//catch Ctrl-C and Kill -9 signals
	signal.Notify(signalCatchingChannel, syscall.SIGINT, syscall.SIGTERM)

	return signalCatchingChannel
}

func activate(data []byte) []byte {
	fmt.Printf("Activating Plugin\n")
	return nil
}

func stop(data []byte) []byte {
	fmt.Printf("Stoping Plugin\n")
	plugin.Stop()
	return nil
}

func Do(data []byte) []byte {
	fmt.Printf("I'm Doing\n")
	return nil
}
