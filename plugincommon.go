package pluginmanager

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// Function to copy a file
func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}

	return
}

// Function to copy a dir
func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := source + "/" + obj.Name()

		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = CopyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

// Parse a version
func parseVersion(s string, width int) int64 {
	strList := strings.Split(s, ".")
	format := fmt.Sprintf("%%s%%0%ds", width)
	v := ""
	for _, value := range strList {
		v = fmt.Sprintf(format, v, value)
	}
	var result int64
	var err error
	if result, err = strconv.ParseInt(v, 10, 64); err != nil {
		fmt.Printf("ugh: parseVersion(%s): error=%s", s, err)
		return 0
	}
	fmt.Printf("parseVersion: [%s] => [%s] => [%d]\n", s, v, result)
	return result
}

// Compare a version of a plugin
func isVersionEqual(start string, end string, current string) bool {

	var startVer int64
	var endVer int64
	var currVer int64

	startVer = parseVersion(start, 4)
	currVer = parseVersion(current, 4)
	if end == "" {
		if startVer-currVer != 0 {
			return false
		}
	} else {
		endVer = parseVersion(end, 4)
		if currVer < startVer || currVer > endVer {
			return false
		}
	}
	return true
}

// Get the name of the function by a function reference
func getFuncName(i interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	dir := filepath.Dir(name)
	// Get the reduced name of the method
	return name[len(dir)+1 : len(name)]
}

// load the config data from the plugin conf file
func loadPluginConfigs(fname string) (PluginConf, error) {
	// open the config file
	configuration := PluginConf{}
	file, err := os.Open(fname)
	defer file.Close()
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

// load the config data from the plugin runtime conf file
func loadRuntimeConfigs(fname string) (RuntimeConf, error) {
	// open the config file
	configuration := RuntimeConf{}
	file, err := os.Open(fname)
	defer file.Close()
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
func saveRuntimeConfigs(fileName string, pluginConf RuntimeConf) error {
	// open the config file
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		return err
	}
	// Encode the data
	encodedData, encodeErr := json.Marshal(pluginConf)
	if encodeErr != nil {
		return encodeErr
	}
	// Write the data to the file
	_, writeErr := file.Write(encodedData)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

// untar a tar file in the alleydog
func untarIt(tarpath string, newpath string) error {

	file, err := os.Open(tarpath)
	if err != nil {
		return err
	}

	sourcefile := tarpath

	defer file.Close()

	var fileReader io.ReadCloser = file

	// just in case we are reading a tar.gz file, add a filter to handle gzipped file
	if strings.HasSuffix(sourcefile, ".gz") {
		if fileReader, err = gzip.NewReader(file); err != nil {
			return err
		}
		defer fileReader.Close()
	}

	tarBallReader := tar.NewReader(fileReader)

	// Extracting tarred files

	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// get the individual filename and extract to the current directory
		filename := filepath.Join(newpath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// handle directory
			err = os.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755 if you prefer

			if err != nil {
				fmt.Printf("Error while creating directory")
				return err
			}

		case tar.TypeReg:
			// handle normal file

			// Get the dir of the filename
			fileDir := filepath.Dir(filename)
			direrr := os.MkdirAll(fileDir, 0755) // or use 0755 if you prefer

			if direrr != nil {
				fmt.Printf("Error while creating directory: %s", fileDir)
				return direrr
			}

			writer, err := os.Create(filename)
			if err != nil {
				fmt.Printf("Error while creating file")

				return err
			}
			io.Copy(writer, tarBallReader)
			err = os.Chmod(filename, os.FileMode(header.Mode))
			if err != nil {
				fmt.Printf("Error while changing file mode")
				return err
			}
			writer.Close()
		default:
			fmt.Printf("Unable to untar type : %c in file %s", header.Typeflag, filename)
		}
	}

	return nil
}

// Get byte from a structure
func getBytes(t interface{}) []byte {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, t)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// Get a interface populated from a byte
func LoadInterface(buf []byte, t interface{}) {
	buffer := bytes.NewBuffer(buf)
	binary.Read(buffer, binary.BigEndian, t)
}
