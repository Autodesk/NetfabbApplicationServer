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
// netfabbtask_pan.go
// Pan Server
//////////////////////////////////////////////////////////////////////////////////////////////////////


package main

import (
	"fmt"
	"time"
)


func RunPanClient (session * NetStorageSession, passedUUID string, Parameters * map[string]string, Results * map[string]string) error {	

	formattedUUID, err := checkUUIDFormat (passedUUID);	
	if (err != nil) {
		return err;
	}

	addLogMessage (session, fmt.Sprintf ("Running Pan Service %s", formattedUUID), LOGTYPE_PANSERVICE, LOGLEVEL_CONSOLE);	

	return nil;
	
}



func RunTaskClient (endpointurl string, worker_name string) {


	session := createEmptySession ();

	handleurl := fmt.Sprintf ("%shandle", endpointurl);

	for true {
		var handlerequest NetTaskHandleRequest;
		handlerequest.Protocol = PROTOCOL_TASKHANDLE;
		handlerequest.Version = PROTOCOL_VERSION;
		handlerequest.Name = "pan";
		handlerequest.Worker = worker_name;
				
		var handlereply NetTaskHandleReply;
		err := sendJSONPostRequest (handleurl, PROTOCOL_TASKHANDLE, &handlerequest, &handlereply);
		if (err != nil) {
			fmt.Println (err);
		} 
		
		if handlereply.UUID != "" {

			var updaterequest NetTaskUpdateRequest;
			updaterequest.Protocol = PROTOCOL_TASKUPDATE;
			updaterequest.Version = PROTOCOL_VERSION;
			updaterequest.WorkerSecret = handlereply.WorkerSecret;
			updaterequest.Results = make(map[string]string);
		
			err = RunPanClient (&session, handlereply.UUID, &handlereply.Parameters, &updaterequest.Results);

			if (err == nil) {
				updaterequest.Status = "SUCCESS";
			} else {
				updaterequest.Status = "ERROR";
			}
						
			
			updateurl := fmt.Sprintf ("%s%s/update", endpointurl, handlereply.UUID);
			
			var updatereply NetTaskHandleReply;
			err = sendJSONPostRequest (updateurl, PROTOCOL_TASKUPDATE, &updaterequest, &updatereply);
			if (err != nil) {
				fmt.Println (err);
			} 
		}
			
		time.Sleep(1000 * time.Millisecond);
	}
	
	
}


