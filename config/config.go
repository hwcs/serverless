package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Open_lambda_home     string
	Redis_network        string
	Redis_address        string
	LoadBalancer_address string
	Registry_address     string
	Use_proxy            string
	Http_proxy           string
	Https_proxy          string
	Base_image           string
}

var Conf Config

func (c *Config) Dump() {
	s, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	log.Printf("CONFIG = %v\n", string(s))
}

func (c *Config) defaults() error {

	return nil
}

func ParseConfig() error {
	Conf.Open_lambda_home = os.Getenv("OPEN_LAMBDA_HOME")
	file := Conf.Open_lambda_home + "/conf/funcstore.json"
	config_raw, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Errorf("ParseConfig, could not open config (%v): %v\n", file, err.Error())
		return err
	}
	var v = make(map[string]string)

	if err := json.Unmarshal(config_raw, &v); err != nil {
		log.Printf("ParseConfig, FILE: %v\n", config_raw)
		fmt.Errorf("ParseConfig, could not parse config (%v): %v\n", file, err.Error())
		return err
	}

	Conf.Redis_network = v["redis_network"]
	Conf.Redis_address = v["redis_host"] + ":" + v["redis_port"]
	Conf.LoadBalancer_address = "http://" + v["loadbalancer_host"] + ":" + v["loadbalancer_port"]
	Conf.Registry_address = v["registry_host"] + ":" + v["registry_port"]
	Conf.Use_proxy = v["use_proxy"]
	Conf.Http_proxy = v["http_proxy"]
	Conf.Https_proxy = v["https_proxy"]
	Conf.Base_image = v["base_image"]

	Conf.Dump()

	return nil
}

