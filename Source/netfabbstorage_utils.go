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
// netfabbstorage_utils.go
// contains Utility functions for the repository
//////////////////////////////////////////////////////////////////////////////////////////////////////


package main

import (
	"net/http"
	"encoding/json"
	"database/sql"
	"io"
	"bytes"
	"regexp"
	"path"
	"errors"
	"io/ioutil"
	"os"
	"fmt"
	"github.com/twinj/uuid"	
	_ "github.com/mattn/go-sqlite3"	
)

//////////////////////////////////////////////////////////////////////////////////////////////////////
// UUID handling functions:
//  - createUUID: creates a new UUID string
//  - checkUUIDFormat: reformats a UUID string to standard format and checks for validity
//////////////////////////////////////////////////////////////////////////////////////////////////////


func createUUID () (string) {
	uuidobject := uuid.NewV4();
	return uuid.Formatter (uuidobject, uuid.FormatCanonical);
}

func checkUUIDFormat (uuidString string) (string, error) {
	uuidobject, err := uuid.Parse(uuidString);
	if (err == nil) {
		return uuid.Formatter (uuidobject, uuid.FormatCanonical), nil;
	} else {
		return "", err
	}
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// userIsValidIdentifier
//  Checks if a userID has valid characters in it
//////////////////////////////////////////////////////////////////////////////////////////////////////

func userIsValidIdentifier (name string) bool {
	var IsValidIdentifier = regexp.MustCompile("^[a-zA-Z0-9_@][a-zA-Z0-9_@]{0,63}$").MatchString
	if (name != "") {
		return IsValidIdentifier (name);
	}
	return false;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// UUID handling functions:
//  - urlCheckRootURL: checks if a url is equal a root end point
//  - parseUUIDURL: parses, if a url is of the form /xxx/<uuid>
//////////////////////////////////////////////////////////////////////////////////////////////////////


func urlCheckRootURL (url string, desiredurl string, isdirectory bool) bool {
	if (isdirectory) {
		if (url == ("/" + desiredurl + "/")) {
			return true;
		}
	}
	
	if (url == ("/" + desiredurl)) {
		return true;
	}
	
	return false;
}

func checkURLPath (url string, desiredurl string) bool {
	desiredpath := "/" + desiredurl;
	
	baselen := len (desiredpath);	
	if (len(url) < baselen) {
		return false;
	}

	suburl := url[0:baselen];
	
	return (suburl == desiredpath);
}



func parseUUIDURL (url string, desiredurl string, desiredadditionalpath string, uuid *string) bool {
	desiredpath := "/" + desiredurl + "/";
	
	baselen := len (desiredpath);	
	if len(url) < (baselen + 36) {
		return false;
	}
	
	suburl := url[0:baselen];
	rawurl := url[baselen:baselen+36];
	additionalurl := url[baselen+36:]
	
	if (suburl == desiredpath) {

		if (desiredadditionalpath != "") {
			bAdditionalCheck := ((additionalurl == "/" + desiredadditionalpath) || (additionalurl == "/" + desiredadditionalpath + "/"));
			if (!bAdditionalCheck) {
				return false;
			}
		}
	
		checkeduuid, err := checkUUIDFormat(rawurl);
		if (err == nil) {
			*uuid = checkeduuid;
			return true;
		}						
	}
	
	return false;
}



func OpenDB (TypeString string, FileName string) (*sql.DB, error) {

	switch (TypeString) {
		case "sqlite":

			db, err := sql.Open("sqlite3", FileName)
			if (err != nil) {
				return nil, errors.New(err.Error () + " (" + FileName + ")");
			}

			return db, nil;
			
		default:
			return nil, errors.New ("Invalid Database Type: " + TypeString);
		
	}
				
}


func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r);
	}
}

func sendJSON (w http.ResponseWriter, replystruct interface{}) error {
	jsondata, err := json.Marshal (replystruct);
	if (err == nil) {
		fmt.Fprintf (w, "%s\n", string (jsondata));
	}					
	
	return err;
}

func parseJSONRequest (r *http.Request, requeststruct interface{}, desiredprotocol string) error {
	jsondata, err := ioutil.ReadAll (r.Body);
	if (err != nil) {
		return err;
	}

	err = json.Unmarshal (jsondata, requeststruct);
	if (err != nil) {
		return err;
	}
	
	headerintf := requeststruct.(NetStorageHeaderInterface);
	protocol := headerintf.GetHeader().Protocol;
	if (protocol != desiredprotocol) {
		return errors.New ("Invalid protocol for end point: " + protocol);
	}
	
	version := headerintf.GetHeader().Version;
	if (version != PROTOCOL_VERSION) {
		return errors.New ("Invalid protocol version for end point: " + version);
	}
		
	return err;
}


func sendError (w http.ResponseWriter, errormsg string) error {
	loguuid := createUUID ();

	var reply NetStorageErrorReply;
	reply.Protocol = PROTOCOL_ERROR;
	reply.Version = PROTOCOL_VERSION;
	reply.ErrorMessage = errormsg;
	reply.LogUUID = loguuid;

	jsondata, err := json.Marshal (&reply);
	if (err == nil) {
		http.Error(w, string (jsondata), http.StatusInternalServerError);		
	}					
	

	return err;
}


func sendFile (w http.ResponseWriter, filename string) error {
	file, err := os.Open(filename);
	if (err == nil) {
		_, err = io.Copy(w, file)
		file.Close()	
	}
	
	return err;
}


func sendJSONPostRequest (url string, desiredprotocol string, requeststruct interface{}, replystruct interface{}) error {

	jsonrequestdata, err := json.Marshal (requeststruct);
	if (err != nil) {
		return err;
	}
	
	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonrequestdata));
	if (err != nil) {
		return err;
	}
	
	defer response.Body.Close();

	if (response.StatusCode == http.StatusNotFound) {
		return errors.New (fmt.Sprintf ("End point not found: %s", url));
	}
	
	if ((response.StatusCode != http.StatusInternalServerError) && (response.StatusCode != http.StatusOK)) {
		return errors.New (fmt.Sprintf ("Invalid request status code: %d", response.StatusCode));
	}
			
	jsondata, err := ioutil.ReadAll (response.Body);
	if (err != nil) {
		return err;
	}

	err = json.Unmarshal (jsondata, replystruct);
	if (err != nil) {
		return err;
	}
	
	headerintf := replystruct.(NetStorageHeaderInterface);
	protocol := headerintf.GetHeader().Protocol;
	if (protocol != desiredprotocol) {
		return errors.New ("Invalid protocol for end point: " + protocol);
	}


	version := headerintf.GetHeader().Version;
	if (version != PROTOCOL_VERSION) {
		return errors.New ("Invalid protocol version for end point: " + version);
	}
				
	return err;
	
}

func getUUIDStorageName (uuid string) string {
	return path.Join (GlobalConfig.Data.Directory, uuid + ".dat");
}
