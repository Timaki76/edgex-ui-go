/*******************************************************************************
 * Copyright © 2017-2018 VMware, Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package controller

import (
	"fmt"
	"log"
	"encoding/json"
	"net/http"
	"github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
	"github.com/edgexfoundry/edgex-ui-go/app/configs"
	"github.com/pelletier/go-toml"
)

const AppServiceConfigurableFileName = "configuration.toml"

func DeployConfigurableProfile(w http.ResponseWriter, r *http.Request){
	configuration := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&configuration)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	client, err := initRegistryClient()
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	configurationTomlTree,err := toml.TreeFromMap(configuration)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	err = client.PutConfigurationToml(configurationTomlTree,true)
	if err != nil{
		log.Printf(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("ok"))
}

func DownloadConfigurableProfile(w http.ResponseWriter, r *http.Request){
	configuration := make(map[string]interface{})
	client, err := initRegistryClient()
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	rawConfiguration,err := client.GetConfiguration(&configuration)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	actual, ok := rawConfiguration.(*map[string]interface{})
	if !ok {
		log.Printf("Configuration from Registry failed type check")
		http.Error(w,"InternalServerError",http.StatusInternalServerError)
		return
	}
	configurationTomlTree,err := toml.TreeFromMap(*actual)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	configurationTomlString,err := configurationTomlTree.ToTomlString()
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-toml;charset=UTF-8")
	w.Header().Set("Content-disposition", "attachment;filename=\""+AppServiceConfigurableFileName+"\"")
	w.Write([]byte(configurationTomlString))
}

func initRegistryClient()(registry.Client,error){
	registryConfig := types.Config{
		Host:            configs.RegistryConf.Host,
		Port:            configs.RegistryConf.Port,
		Type:            configs.RegistryConf.Type,
		Stem:            configs.RegistryConf.ConfigRegistryStem,
		ServiceKey:      configs.RegistryConf.ServiceKey,
	}
	client, err := registry.NewRegistryClient(registryConfig)
	if err != nil {
		return nil,fmt.Errorf("Connection to Registry could not be made: %v", err)
	}
	if !client.IsAlive() {
		return nil,fmt.Errorf("Registry (%s) is not running", registryConfig.Type)
	}
	return client,nil
}