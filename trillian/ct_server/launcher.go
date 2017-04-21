package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// "-log_config": "/config/config.json",
// , "--target", "/main", "--config", "/config/ct.json"

const TRAMPOLINE_CONFIG_PATH = "/config/ct_params.json"
const TEMPLATE_CONFIG_PATH = "/config/ct_config.json"
const GENERATED_CONFIG_PATH = "/log_config.json"

func main() {
	// See if we've already generated the config
	if _, err := os.Stat(GENERATED_CONFIG_PATH); os.IsNotExist(err) {
		// Generate a new config
		log.Println("Generating a new configuration")

		// Figure out where the log server is
		var trampolineArgs map[string]interface{}
		if configData, err := ioutil.ReadFile(TRAMPOLINE_CONFIG_PATH); err != nil {
			log.Fatal("Could not find trampoline config", err)
		} else if err := json.Unmarshal(configData, &trampolineArgs); err != nil {
			log.Fatal("Could not parse trampoline config", err)
		}

		logServerAddrVal, ok := trampolineArgs["-log_rpc_server"]
		if !ok {
			logServerAddrVal, ok = trampolineArgs["--log_rpc_server"]
		}

		if !ok {
			log.Fatal("Could not figure out log server address")
		}

		logServerAddr, ok := logServerAddrVal.(string)
		if !ok {
			log.Fatal("Log server address should be a string, not ", logServerAddrVal)
		}

		// Block until the log is actually available
		for {
			_, err := net.Dial("tcp", logServerAddr)
			if err != nil {
				log.Println("Could not contact log server ", err)
				time.Sleep(time.Duration(1+rand.Intn(10)) * time.Second)
				continue
			}
			break
		}

		// Create the tree, and get the tree ID from the output
		cmd := exec.Command("/createtree", []string{
			"-admin_server", logServerAddr,
			"-pem_key_path", "/config/privkey.pem",
			"-pem_key_password", "towel",
			"-signature_algorithm", "ECDSA",
		}...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal("Could not create new tree", string(output), err)
		}

		logID, err := strconv.Atoi(strings.TrimSpace(string(output)))
		if err != nil {
			log.Fatal(err)
		}

		// Read in the template config and have everyone use our config
		var template []map[string]interface{}
		if templateData, err := ioutil.ReadFile(TEMPLATE_CONFIG_PATH); err != nil {
			log.Fatal("Could not find template config", err)
		} else if err := json.Unmarshal(templateData, &template); err != nil {
			log.Fatal("Could not parse template config", err)
		}

		for i := range template {
			template[i]["LogID"] = logID
		}

		data, err := json.Marshal(template)
		if err != nil {
			log.Fatal("Could not marshal new config", err)
		}
		if err := ioutil.WriteFile(GENERATED_CONFIG_PATH, data, os.ModePerm); err != nil {
			log.Fatal("Could not write new config", err)
		}
		log.Println(string(data))
	} else {
		log.Println("Using existing configuration")
	}

	// Launch the trampoline, with the config parameter
	cmd := exec.Command("/trampoline", []string{
		"--target", "/main",
		"--config", TRAMPOLINE_CONFIG_PATH,
		"--",
		"--log_config", GENERATED_CONFIG_PATH,
	}...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	// TODO: also pass on other cli arguments
}
