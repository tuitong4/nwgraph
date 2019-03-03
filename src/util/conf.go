package util

import (
	"encoding/json"
	io "io/ioutil"
)

type Config struct {
	Url       string `json:"url"`
	SaveBatch int64  `json:"savebatch"`
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
