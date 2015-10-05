package GoPlug

import (
	"encoding/json"
	"fmt"
	"os"
)

type PluginConf struct {
	Name      string
	NameSpace string
	Url       string
	Sock      string
	LazyLoad  bool
}

// load the config data from the file
func loadConfigs(fname string) (PluginConf, error) {
	// open the config file
	configuration := PluginConf{}
	file, err := os.Open(fname)
	if err != nil {
		return configuration, err
	}
	// load the config from file
	decoder := json.NewDecoder(file)
	loaderr := decoder.Decode(&configuration)
	if loaderr != nil {
		return configuration, loaderr
	}

	return configuration, nil
}

// save the config data to the file
func saveConfigs(fileName string, pluginConf PluginConf) error {
	// open the config file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	// Encode the data
	encodedData, encodeErr := json.Marshal(pluginConf)
	if encodeErr != nil {
		return encodeErr
	}
	// Write the data to the file
	dataSize, writeErr := file.Write(encodedData)
	if writeErr != nil {
		return writeErr
	}

	fmt.Printf("wrote %d bytes\n", dataSize)

	return nil
}
