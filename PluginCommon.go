package GoPlug

import (
	"encoding/json"
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
