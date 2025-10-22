package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

var configFile = "config.yaml"
var currentIpFile = "current_ip.txt"
var logFile = "ip_update.log"
var logging = true

type LogLevel int

const (
	LogAll LogLevel = iota
	LogWarning
	LogError
)

var logLabels = map[string]LogLevel{
	"all":     LogAll,
	"warning": LogWarning,
	"error":   LogError,
}

var logLevel = LogError

type Apikeys struct {
	ApiKey    string
	SecretKey string
}

type Config struct {
	Ipapi   Ipapi    `yaml:"ipapi"`
	Dnsapi  Dnsapi   `yaml:"dnsapi"`
	Domains []Domain `yaml:"domains"`

	IpFile       string `yaml:"ipfile"`
	LogFile      string `yaml:"logfile"`
	Logging      bool   `yaml:"logging"`
	LoggingLevel string `yaml:"logging_level"`
}

type Ipapi struct {
	Address []string `yaml:"address"`
}

type Dnsapi struct {
	UpdateEndpoint string `yaml:"update_endpoint"`
	ReadEndpoint   string `yaml:"read_endpoint"`
	Apikey         string `yaml:"apikey"`
	Secretkey      string `yaml:"secretkey"`
}

type Domain struct {
	Domain    string `yaml:"domain"`
	Subdomain string `yaml:"subdomain"`
	Wildcard  bool   `yaml:"wildcard"`
	Id        string `yaml:"id"`
}

type ReadResponse struct {
	Status  string   `json:"status"`
	Records []Record `json:"records"`
}

type Record struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Ttl     string `json:"ttl"`
	Prio    string `json:"prio"`
	Notes   string `json:"notes"`
}

func readIpFile() (string, error) {
	ip, err := os.ReadFile(currentIpFile)
	if err != nil {
		return "", fmt.Errorf("Failed to read IP file (%w): %w", currentIpFile, err)
	}

	return string(ip), nil
}

func writeIpFile(ip string) error {
	err := os.WriteFile(currentIpFile, []byte(ip), 0644)
	if err != nil {
		return fmt.Errorf("Failed to write IP file (%w): %w", currentIpFile, err)
	}
	fmt.Printf("IP saved to %w\n", currentIpFile)

	return nil
}

func (conf *Config) readConfig() (*Config, error) {
	fmt.Println("Reading config YAML")
	defer fmt.Println("Done reading config")

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read YAML (%w): %w", configFile, err)
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse YAML (%w): %w", configFile, err)
	}

	if conf.IpFile != "" {
		currentIpFile = conf.IpFile
	}
	if conf.LogFile != "" {
		logFile = conf.LogFile
	}

	var ok bool
	logLevel, ok = logLabels[conf.LoggingLevel]
	if !ok {
		logLevel = LogError
	}

	logging = conf.Logging

	/* 	apiKey := os.Getenv("DNS_APIKEY")
	   	apiSecret := os.Getenv("DNS_APISECRET")

	   	if apiKey != "" && apiSecret != "" {
	   		conf.Dnsapi.Apikey = apiKey
	   		conf.Dnsapi.Secretkey = apiSecret
	   	}

	   	if conf.Dnsapi.Apikey == "" || conf.Dnsapi.Secretkey == "" {
	   		return nil, fmt.Errorf("API key or secret not set!")
	   	} */

	return conf, nil
}

func (res *ReadResponse) readQuery(conf Config) error {
	fmt.Println("Reading DNS records from API")
	defer fmt.Println("Read complete")

	if len(conf.Domains) == 0 {
		return fmt.Errorf("No domains configured")
	}

	post_data := map[string]string{
		"apikey":       conf.Dnsapi.Apikey,
		"secretapikey": conf.Dnsapi.Secretkey,
	}

	fmt.Println(post_data)

	json_data, err := json.Marshal(post_data)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON data: %w", err)
	}

	apiendpoint := conf.Dnsapi.ReadEndpoint + "/" + conf.Domains[0].Domain

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Post(apiendpoint, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return fmt.Errorf("Error in making POST query")
	}

	defer resp.Body.Close()

	fmt.Println(apiendpoint)
	fmt.Println(string(json_data))

	fmt.Println(resp.Status)

	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return fmt.Errorf("Failed to decode JSON response: %w", err)
	}

	fmt.Printf("Received %d records\n", len(res.Records))

	return nil
}

func updateDns(conf *Config, ip string) error {

	fmt.Println("Updating IP...")
	defer fmt.Println("Update complete")

	if len(conf.Domains) == 0 {
		return fmt.Errorf("No domains configured")
	}

	post_data := map[string]string{
		"apikey":       conf.Dnsapi.Apikey,
		"secretapikey": conf.Dnsapi.Secretkey,
		"content":      ip,
		"ttl":          "600",
	}

	fmt.Println(post_data)

	json_data, err := json.Marshal(post_data)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON data: %w", err)
	}

	apiendpoint := conf.Dnsapi.UpdateEndpoint + "/" + conf.Domains[0].Domain

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Post(apiendpoint, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return fmt.Errorf("Failed to make POST")
	}

	defer resp.Body.Close()

	fmt.Println(apiendpoint)
	fmt.Println(string(json_data))

	fmt.Println(resp.Status)

	var res map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return fmt.Errorf("Failed to decode JSON response: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to update IP: %w", resp.Status)
	}

	return nil

}

func checkCurrentIp(conf *Config) (string, error) {

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(conf.Ipapi.Address[0])
	if err != nil {
		return "", fmt.Errorf("Failed to GET IP: %w", err)
	}
	defer resp.Body.Close()

	var res map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("Failed to decode JSON response to current IP: %w", err)
	}

	ip, ok := res["ip"].(string)
	if !ok {
		return "", fmt.Errorf("Field IP not found in response or not a string")
	}

	return ip, nil
}

func (keys *Apikeys) getApiKeys() error {
	keys.ApiKey = os.Getenv("DNS_APIKEY")
	keys.SecretKey = os.Getenv("DNS_APISECRET")

	/* 	if keys.ApiKey != "" && keys.SecretKey != "" {
		conf.Dnsapi.Apikey = apiKey
		conf.Dnsapi.Secretkey = apiSecret
	} */

	if keys.ApiKey == "" || keys.SecretKey == "" {
		return fmt.Errorf("API key or secret not set!")
	}

	return nil
}

func printAndLog(text string, level LogLevel) {

	if logging {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		writer := io.MultiWriter(os.Stderr, file)
		log.SetOutput(writer)
	}

	switch logLevel {
	case LogAll:
		// log all
		log.Println(text)

	case LogWarning:
		// warnings
		if level >= LogWarning {
			log.Println(text)
		}
	case LogError:
		// errors only
		if level == LogError {
			log.Println(text)
		}
	default:
		// errors only
		if level == LogError {
			log.Println(text)
		}
	}
}

func main() {

	var conf Config
	_, err := conf.readConfig()
	if err != nil {
		printAndLog(err.Error(), LogError)
		os.Exit(1)
	}

	var keys Apikeys
	err = keys.getApiKeys()
	if err != nil {
		printAndLog(err.Error(), LogError)
		os.Exit(1)
	}

	//	apiKey := os.Getenv("DNS_APIKEY")
	//	apiSecret := os.Getenv("DNS_APISECRET")

	//	if apiKey == "" && conf.Dnsapi.Apikey != "" {
	//		apiKey = conf.Dnsapi.Apikey
	//	} else {
	//		fmt.Println("No API key found!")
	//		os.exit(1)
	//	}

	//	if apiSecret == "" && conf.Dnsapi.Secretkey != "" {
	//		apiSecret = conf.Dnsapi.Secretkey
	//	} else {
	//		fmt.Println("No API secret key found!")
	//		os.exit(1)
	//	}

	//	fmt.Println(conf.Ipapi.Address[0])
	//	fmt.Println(conf.Dnsapi.UpdateEndpoint)

	//	fmt.Println(conf.Dnsapi.ReadEndpoint)

	fmt.Println("Reading DNS entries")

	prevIp, err := readIpFile()
	if err != nil {
		printAndLog(err.Error(), LogWarning)
	}

	fmt.Println(prevIp)

	currentIp, err := checkCurrentIp(&conf)
	if err != nil {
		printAndLog(err.Error(), LogError)
		os.Exit(1)
	}

	fmt.Println(currentIp)

	if prevIp != currentIp {
		err = updateDns(&conf, currentIp)
		if err != nil {
			printAndLog(err.Error(), LogError)
			os.Exit(1)
		}

		err = writeIpFile(currentIp)
		if err != nil {
			printAndLog(err.Error(), LogError)
		}
	}

	//	var read ReadResponse
	//	read.readQuery(&conf)

	//	for _, rec :=  range read.Records {
	//		fmt.Printf("- %s (%s): %s\n", rec.Name, rec.Type, rec.Content)
	//	}
}
