package main

import (
	"com.ss/goplugin/PluginImpl"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	config := PluginImpl.PluginImplConf{"PluginA.conf", activate, stop}
	plugin, err := PluginImpl.PluginInit(config)
	if err != nil {
		fmt.Printf("Plugin Init Error: %s\n", err)
		return
	}
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
	fmt.Printf("Activating Plugin")
	return nil
}

func stop(data []byte) []byte {
	fmt.Printf("Stoping Plugin")
	return nil
}
