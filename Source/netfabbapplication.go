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
// netfabbstorageserver.go
// contains HTTP Handler functions for every REST endpoint
//////////////////////////////////////////////////////////////////////////////////////////////////////


package main

import (
	"net/http"
	"fmt"
	"log"
	"errors"
	"crypto/sha1"
	"strings"
)


var GlobalConfig ConfigDefinition;


func StorageSessionGetHashedSalt (UserID string) (string, string, error) {

	// Default salt is the global access key
	Salt := GlobalConfig.Authentication.Global.Salt;
	Passphrase := GlobalConfig.Authentication.Global.Passphrase;
		
	// Check if we know the user
	for j := 0; j < len(GlobalConfig.Authentication.NamedUsers); j++ {
		user := GlobalConfig.Authentication.NamedUsers[j];
		
		if (user.UserID == UserID) {
			Salt = user.Salt;
			Passphrase = user.Passphrase;
		}
		
	}
	
	
	if (Salt != "") {
	
		hash := sha1.New()
		hash.Write([]byte(Salt))
		calculatedHash := hash.Sum (nil);
		hashString := fmt.Sprintf ("%x", calculatedHash);
		
		return Passphrase, hashString, nil;
		
	} else {
		
		return Passphrase, "", nil;
	}

}


func StorageSessionNew (w http.ResponseWriter, r *http.Request) error {
	
	// Parse JSON request
	var request NetStorageCreateSessionRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_NEWSESSION);
	if (err != nil) {
		return err;
	}
	
		
	newSession, err := createNewSessionInDB (request.UserID);
	if (err != nil) {
		return err;
	}
	
	addLogMessage(&newSession, fmt.Sprintf("Created Session for user \"%s\"...", request.UserID), LOGTYPE_SYSTEM, LOGLEVEL_CONSOLE);


	_, encryptedSalt, err := StorageSessionGetHashedSalt (request.UserID);
	if (err != nil) {
		return err;
	}
	
	// Send reply JSON	
	var reply NetStorageCreateSessionReply;
	reply.Protocol = PROTOCOL_NEWSESSION;
	reply.Version = PROTOCOL_VERSION;
	reply.SessionUUID = newSession.UUID;	
	reply.AuthType = "saltedhash";
	reply.UserID = request.UserID;
	reply.Salt = encryptedSalt;
		
	return sendJSON (w, &reply);			
	
	
}


func StorageSessionAuth (w http.ResponseWriter, r *http.Request) error {
	
	// Parse JSON request
	var request NetStorageAuthenticateSessionRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_AUTHSESSION);
	if (err != nil) {
		return err;
	}
	
	
	if GlobalConfig.Authentication.Type != "passphrase" {
		return errors.New ("Unknown authentication method");
	}

	if request.AuthType != "saltedhash" {
		return errors.New ("Invalid authentication method");
	}

	
	formattedUUID, err := checkUUIDFormat (request.SessionUUID);
	if (err != nil) {
		return err;
	}

	
	UserID, err := RetrieveUserIDFromSession (request.SessionUUID);
	if (err != nil) {
		return err;
	}

	passphrase, _, err := StorageSessionGetHashedSalt (UserID);
	if (err != nil) {
		return err;
	}
	
	stringToHash := "NETFABB" + formattedUUID + passphrase;
	if (err != nil) {
		return err;
	}

	hash := sha1.New()
	hash.Write([]byte(stringToHash))
	calculatedHash := hash.Sum (nil);
	
	calculatedKey := fmt.Sprintf ("%x", calculatedHash);
	suppliedKey := strings.TrimSpace (strings.ToLower (request.AuthKey));

	if (calculatedKey != suppliedKey) {
		return errors.New ("Authentication failed - invalid connection key!");
	}
	
	//log.Println ("key1: ", suppliedKey);
	//log.Println ("key2: ", calculatedKey);
	
	Token, err := AuthenticateSessionInDB (request.SessionUUID, request.AuthType, request.AuthKey);
	if (err != nil) {
		return err;
	}
		
		
	// Send reply JSON	
	var reply NetStorageAuthenticateSessionReply;
	reply.Protocol = PROTOCOL_AUTHSESSION;
	reply.Version = PROTOCOL_VERSION;
	reply.SessionUUID = request.SessionUUID;	
	reply.Token = Token;
		
	return sendJSON (w, &reply);			
	
	
}



//////////////////////////////////////////////////////////////////////////////////////////////////////
// Auth handler
//////////////////////////////////////////////////////////////////////////////////////////////////////

func AuthHandler (w http.ResponseWriter, r *http.Request) (bool, error) {

	url := r.URL.Path;

	if (r.Method == "POST") {		
		if urlCheckRootURL (url, "session/new", true) {
			err := StorageSessionNew (w, r);
			return true, err;
		}

		if urlCheckRootURL (url, "session/auth", true) {
			err := StorageSessionAuth (w, r);
			return true, err;
		}
		
	}
	
	return false, nil;
	
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// RESTHandler
// handles all requests to the storage REST endpoint
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RESTHandler (w http.ResponseWriter, r *http.Request) {

	// URL
	url := r.URL.Path;
	
	// Session authentication needs to be handled separately from business logic
	if checkURLPath (url, "session/") {	
		success, err := AuthHandler (w, r);
		if (err != nil) {
			sendError (w, err.Error ());			
			log.Println (err);
		}
				
		if (!success) {
			http.NotFound (w, r);
		}					
				
	} else {
		
		AuthHeader := r.Header.Get ("Authorization");
		
		if len (AuthHeader) < 8 {
			http.Error (w, "Invalid authorization header", http.StatusNetworkAuthenticationRequired );
			return
		}
		
		if (AuthHeader[0:7] != "Bearer ") {
			http.Error (w, "Invalid authorization token", http.StatusNetworkAuthenticationRequired );
			return
		}		
		
		Token := AuthHeader[7:]
			
		// Authentication
		session, err := retrieveSessionByToken (Token, GlobalConfig.Authentication.DurationOfSessions);
		if (err != nil) {
			http.Error (w, "could not authenticate", http.StatusForbidden);
			addLogMessage (&session, "Session Error: " + url, LOGTYPE_SYSTEM, LOGLEVEL_CONSOLE);			
			return
		}
		
		
		// Handle request		
		addLogMessage (&session, "Retrieved request: " + url, LOGTYPE_SYSTEM, LOGLEVEL_DBONLY);
			
		// Create DB Connection
		db, err := OpenDB (GlobalConfig.Database.Type, GlobalConfig.Database.FileName);
		defer db.Close()
		if (err == nil) {
		
			success := false;
			
			if checkURLPath (url, "data/") {
				success, err = DataHandler (db, &session, w, r);
			}
	/*		
			-- disabled for now
			if checkURLPath (url, "orm/") {
				success, err = ORMHandler (db, &session, w, r);
			} 
			
			*/
			if checkURLPath (url, "tasks/") {
				success, err = TaskHandler (db, &session, w, r);
			}
			
			
			if (err != nil) {
				sendError (w, err.Error ());			
				log.Println (err);
			}
			
			
			if (!success) {
				http.NotFound (w, r);
			}	
		
					
		} else {
			sendError (w, "Could not open Database.");			
			log.Println (err);
		}
	
	}
		
}



//////////////////////////////////////////////////////////////////////////////////////////////////////
// main function
//////////////////////////////////////////////////////////////////////////////////////////////////////

func startAppServer (ConfigFileName string, runBlocking bool) (error) {	

	
	// Loading Database Schemas
	/*
	
	Disabled for now
	
	log.Println(fmt.Sprintf("Loading Database Schemas.."));		
	count, err := ORMInitialiseSchema ("netfabbormschemas.json");
	if (err != nil) {
		return err;
	}		
	log.Println(fmt.Sprintf("   found %d tables..", count));		
	
	
	*/

	// Handle REST endpoint				

	var err error;
	
	if (ConfigFileName != "") {
		log.Println(fmt.Sprintf("Loading config file %s...", ConfigFileName));		
		GlobalConfig, err = LoadConfig (ConfigFileName);
	} else {
		GlobalConfig, err = LoadConfigFromRegistry ();
	}

	if (err != nil) {
		return err;
	}		
	
	host := GlobalConfig.Server.Host;
	port := GlobalConfig.Server.Port;
		
	// Get Endpoint URL
	// workername := CONFIG_WORKERNAME;
	//task_endpointurl := fmt.Sprintf ("http://%s:%d/tasks/", host, port);

	// Run Task handlers -- disabled for now!
	//if CONFIG_RUNPANSERVICE {
		//go RunTaskClient (task_endpointurl, workername);
	//	}

	// Initialize Logging and Session handling
	log.Println(fmt.Sprintf("Initializing Log Database.."));		
	err, dbfilename := initSessionDB (GlobalConfig.Log.Prefix);
	if (err != nil) {
		log.Println(fmt.Sprintf("Error while creating %s..", dbfilename));		
		return err;
	}		
	
	session := createEmptySession ();
	
	addLogMessage(&session, fmt.Sprintf("Logging to %s..", dbfilename), LOGTYPE_SYSTEM, LOGLEVEL_CONSOLE);
	addLogMessage(&session, fmt.Sprintf("Listening on host %s, port %d..", host, port), LOGTYPE_SYSTEM, LOGLEVEL_CONSOLE);
    
	http.Handle ("/", makeHandler (RESTHandler));
	
	error_channel := make (chan error)
	
	switch (GlobalConfig.HTTPS.Type) {
		case "tls":
			go func () {
				err := http.ListenAndServeTLS(fmt.Sprintf("%s:%d", host, port), GlobalConfig.HTTPS.Certificate, GlobalConfig.HTTPS.PrivateKey, nil)
				error_channel <- err;
			} ()
		
		case "none", "":
			go func () {
				err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
				error_channel <- err;
			} ()
			
		default:
			error_channel <- errors.New ("Invalid HTTPS type: " + GlobalConfig.HTTPS.Type);
	}
	

	if (runBlocking) {
		err = <- error_channel;
		if (err != nil) {
			return err;
		}		
	}

	return nil;
}


