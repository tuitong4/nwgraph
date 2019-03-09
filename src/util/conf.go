package util

import (
	"encoding/json"
	io "io/ioutil"
)

type Config struct {
	Url         string `json:"url"`
	SaveBatch   int64  `json:"savebatch"`
	LogFile     string `json:"logfile"`
	NeoServer   string `json:"neoserver"`
	NeoUser     string `json:"neouser"`
	NeoPassword string `json:"neopassword"`
}

func NewConfig(file string) (*Config, error) {
	c := &Config{Url: ""}

	data, err := io.ReadFile(file)
	if err != nil {
		return nil, err
	}
	datajson := []byte(data)

	err = json.Unmarshal(datajson, c)

	if err != nil {
		return nil, err
	}
	return c, nil
}
