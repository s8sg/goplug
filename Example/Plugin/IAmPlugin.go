package main

import (
	"fmt"
	GoPlug "github.com/swarvanusg/GoPlug"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	config := GoPlug.PluginImplConf{"./", "Do", "Test", "", false, activate, stop}
	plugin, err := GoPlug.PluginInit(config)
	if err != nil {
		fmt.Printf("Plugin Init Error: %s\n", err)
		return
	}
	plugin.RegisterMethod("Do", Do)
	plugin.Start()
	shutdownChannel := makeShutdownChannel()

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
	return nil
}

func Do(data []byte) []byte {
	fmt.Printf("I'm Doing\n")
	return nil
}
