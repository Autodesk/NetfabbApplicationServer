/*++

Copyright (C) 2015 Autodesk Inc. (Original Author)

All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

--*/

//////////////////////////////////////////////////////////////////////////////////////////////////////
// netfabbstorage_config.go
// Configuration constants
//////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	
	"golang.org/x/sys/windows/registry" 
)


const CONFIG_SESSIONDBPREFIX = "./logs/log_";
const CONFIG_DEFAULTPORT = 8650;
const CONFIG_DEFAULTHOST = "localhost";

const CONFIG_NAME = "Autodesk Netfabb Application Server";
const CONFIG_VERSION = "v1.0.0";
const CONFIG_DBNAME = "./netfabbapplicationserver.db";

const CONFIG_DEFAULTDATADIRECTORY = "./data/";

const CONFIG_WORKERNAME = "ApplicationServer";
const CONFIG_RUNPANSERVICE = false;


type ConfigDefinitionServer struct {
	XMLName xml.Name `xml:"server"`
	Host string `xml:"host,attr"`
	Port int `xml:"port,attr"`
}

type ConfigDefinitionLog struct {
	XMLName xml.Name `xml:"log"`
	Prefix string `xml:"prefix,attr"`
}

type ConfigDefinitionDatabase struct {
	XMLName xml.Name `xml:"database"`
	Type string `xml:"type,attr"`
	FileName string `xml:"filename,attr"`
}


type ConfigDefinitionData struct {
	XMLName xml.Name `xml:"data"`
	Directory string `xml:"directory,attr"`
}

type ConfigDefinitionHTTPS struct {
	XMLName xml.Name `xml:"https"`
	Type string `xml:"type,attr"`
	Certificate string `xml:"certificate,attr"`
	PrivateKey string `xml:"privatekey,attr"`
}


type ConfigDefinitionAuthenticationNamedUser struct {
	XMLName xml.Name `xml:"nameduser"`
	UserID string `xml:"id,attr"`
	Passphrase string `xml:"passphrase,attr"`
	Salt string `xml:"salt,attr"`
}

type ConfigDefinitionAuthenticationGlobal struct {
	XMLName xml.Name `xml:"global"`
	Passphrase string `xml:"passphrase,attr"`
	Salt string `xml:"salt,attr"`
}


type ConfigDefinitionAuthentication struct {
	XMLName xml.Name `xml:"authentication"`
	Type string `xml:"type,attr"`
	DurationOfSessions int `xml:"sessionduration,attr"`

	NamedUsers []ConfigDefinitionAuthenticationNamedUser `xml:"nameduser"`
	Global ConfigDefinitionAuthenticationGlobal `xml:"global"`	
}


type ConfigDefinition struct {
	XMLName xml.Name `xml:"config"`
	XMLNameSpace string `xml:"xmlns,attr"`
	Server ConfigDefinitionServer `xml:"server"`
	Log ConfigDefinitionLog `xml:"log"`
	Database ConfigDefinitionDatabase `xml:"database"`
	Data ConfigDefinitionData `xml:"data"`
	HTTPS ConfigDefinitionHTTPS `xml:"https"`
	Authentication ConfigDefinitionAuthentication `xml:"authentication"`
	
}



func LoadConfig (FileName string) (ConfigDefinition, error) {

	var config ConfigDefinition;
	config.Server.Host = CONFIG_DEFAULTHOST;
	config.Server.Port = CONFIG_DEFAULTPORT;
	config.Log.Prefix = CONFIG_SESSIONDBPREFIX;
	
	file, err := os.Open(FileName);
	if (err != nil) {
		return config, err
	}
	
	defer file.Close();

	bytes, err := ioutil.ReadAll (file);
	if (err != nil) {
		return config, err
	}
	
	err = xml.Unmarshal(bytes, &config)
	if (err != nil) {
		return config, err
	}

	return config, nil 
	
}



func LoadConfigFromRegistry () (ConfigDefinition, error) {

	var config ConfigDefinition;
	
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Autodesk\Netfabb Application Server`, registry.QUERY_VALUE)
	if err != nil {
		return config, err;
	}
	
	defer key.Close()

	
	
	ConfigXMLName, _, err := key.GetStringValue("ConfigXML")
	if err != nil {
		return config, err;
	}
	
	return LoadConfig (ConfigXMLName)
}



func SaveConfigPathToRegistry (ConfigXMLName string) (error) {

	
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SOFTWARE\Autodesk\Netfabb Application Server`, registry.SET_VALUE)
	if err != nil {
		return err;
	}
	
	defer key.Close()

	
	
	err = key.SetStringValue("ConfigXML", ConfigXMLName)
	
	return err
}


