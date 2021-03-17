package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

//SaveCfg Saves the config file c to the file fPath as a YAML file
//TODO:
func SaveCfg(c interface{}, fPath string) error {
	f, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	defer f.Close()

	conf, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	fmt.Println(string(conf))
	f.Write(conf)

	return nil
}

//LoadCfg Reads the file on fPath as a YAML format and unmarshals data into conf.
//c should be the default configuration file
//In the case that the designated config file does not exist, this function will attempt to create it using conf as a template
//TODO:
func LoadCfg(conf interface{}, fPath string) error {

	readChunkSize := 2048

	f, err := os.OpenFile(fPath, os.O_RDONLY&os.O_CREATE, os.ModePerm)

	if err != nil {
		if os.IsNotExist(err) {
			return SaveCfg(conf, fPath)
		}
		return err
	}

	defer f.Close()

	data := make([]byte, readChunkSize)

	n, err := f.Read(data)
	if err != nil {
		return err
	}

	if n == readChunkSize {
		exData := make([]byte, readChunkSize)

		// Keep reading in Chunks of size 'readChunkSize' until file is exhausted
		for n == readChunkSize {
			n, err = f.Read(exData)
			if err == io.EOF {
				break
			}
			// shrink buffer to actual size of data to avoid yaml read errors
			exData = exData[:n-1]
			data = append(data, exData...)
		}
	} else {
		// shrink buffer to actual size of data to avoid yaml read errors
		data = data[:n-1]
	}

	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return err
	}
	return nil
}
