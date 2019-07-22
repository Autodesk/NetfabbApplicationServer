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
// netfabbtask_handler.go
// Handles netfabb Tasks
//////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"net/http"
	"fmt"
	"time"
	"errors"
	"encoding/json"
	"database/sql"	
)


//////////////////////////////////////////////////////////////////////////////////////////////////////
// Create a new task
//////////////////////////////////////////////////////////////////////////////////////////////////////


func TaskNewHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) error {
	uuid := createUUID ();

	addLogMessage (session, fmt.Sprintf ("Create new Task %s", uuid), LOGTYPE_TASK_NEW, LOGLEVEL_CONSOLE);	

	// Parse JSON request
	var request NetTaskNewRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_TASKNEW);
	if (err != nil) {
		return err;
	}		
	
	
	jsondata, err := json.Marshal (&request.Parameters);
	if (err != nil) {
		return err;
	}
	
	parameters := string (jsondata);	
	
	timestamp := time.Now().Format(time.RFC3339);
	status := "NEW";
	taskname := request.Name;
	
	if (taskname == "") {
		return errors.New ("Invalid task name");
	}
	
	addLogMessage (session, fmt.Sprintf ("Parameters: %s", parameters), LOGTYPE_TASK_NEW, LOGLEVEL_DBONLY);		
		
	query := fmt.Sprintf ("INSERT INTO netstorage_tasks (uuid, taskname, status, parameters, timestamp, transactionuuid) VALUES (?, ?, ?, ?, ?, ?)");
		
	statement, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}
	
	_, err = statement.Exec (uuid, taskname, status, parameters, timestamp, uuid);
	if (err != nil) {
		return err;
	}
			
	var reply NetTaskNewReply;
	reply.Protocol = PROTOCOL_TASKNEW;
	reply.Version = PROTOCOL_VERSION;
	reply.UUID = uuid;
	return sendJSON (w, &reply);			
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// Clear all open tasks
//////////////////////////////////////////////////////////////////////////////////////////////////////


func TaskClearHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) error {

	transactionUUID := createUUID ();

	addLogMessage (session, fmt.Sprintf ("Clearing all tasks (transaction UUID: %s)", transactionUUID), LOGTYPE_TASK_CLEAR, LOGLEVEL_CONSOLE);	

	// Parse JSON request
	var request NetTaskClearRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_TASKCLEAR);
	if (err != nil) {
		return err;
	}		
		
	query := fmt.Sprintf ("UPDATE netstorage_tasks SET status=?, transactionuuid=? WHERE status=?");
		
	statement, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}
	
	res, err := statement.Exec ("CANCELED", transactionUUID, "NEW");
	if (err != nil) {
		return err;
	}
		
	count, err := res.RowsAffected();
	if (err != nil) {
		return err;
	}

		
	var reply NetTaskClearReply;
	reply.Protocol = PROTOCOL_TASKCLEAR;
	reply.Version = PROTOCOL_VERSION;
	reply.Count = int (count);
	return sendJSON (w, &reply);			
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// Handle a task and mark it "INPROCESS"
//////////////////////////////////////////////////////////////////////////////////////////////////////

func TaskHandleHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) error {

	transactionUUID := createUUID ();
	workerSecret := transactionUUID;

	addLogMessage (session, fmt.Sprintf ("handling task (transaction UUID: %s)", transactionUUID), LOGTYPE_TASK_HANDLE, LOGLEVEL_DBONLY);	

	// Parse JSON request
	var request NetTaskHandleRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_TASKHANDLE);
	if (err != nil) {
		return err;
	}		

	worker := request.Worker;
	taskname := request.Name;
	addLogMessage (session, fmt.Sprintf ("Task name: %s, worker: %s", taskname, worker), LOGTYPE_TASK_HANDLE, LOGLEVEL_DBONLY);	
	
	if (taskname == "") {
		return errors.New ("Invalid task name");
	}

	query := fmt.Sprintf ("UPDATE netstorage_tasks SET status=?, transactionuuid=?, worker=?, workersecret=? WHERE uuid IN (SELECT uuid FROM netstorage_tasks WHERE (status=? or status=?) AND taskname=? ORDER BY timestamp DESC LIMIT 1)");
		
	statement1, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}
	
	defer statement1.Close();
	
	_, err = statement1.Exec ("INPROCESS", transactionUUID, worker, workerSecret, "NEW", "RETURNED", taskname);
	if (err != nil) {
		return err;
	}
	
	query = fmt.Sprintf ("SELECT uuid, taskname, parameters FROM netstorage_tasks WHERE transactionuuid=?");
		
	statement2, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}

	defer statement2.Close();
	
	rows, err := statement2.Query (transactionUUID);
	if (err != nil) {
		return err;
	}
	
	defer rows.Close();
	
	resultuuid := "";
	parameters := "";
	
	if (rows.Next()) {
		err = rows.Scan (&resultuuid, &taskname, &parameters);
		if (err != nil) {
			return err;
		}
	
		if (rows.Next()) {
			return errors.New("Duplicate tasks locked in request!");		
		}			
		addLogMessage (session, fmt.Sprintf ("Task retrieved: taskname %s, resultuuid: %s", taskname, resultuuid), LOGTYPE_TASK_HANDLE, LOGLEVEL_CONSOLE);		
		addLogMessage (session, fmt.Sprintf ("  Parameters: %s", parameters), LOGTYPE_TASK_HANDLE, LOGLEVEL_DBONLY);		
		
	} else {
		addLogMessage (session, fmt.Sprintf ("  no task in queue"), LOGTYPE_TASK_HANDLE, LOGLEVEL_DBONLY);		
	}
	

	var reply NetTaskHandleReply;
	reply.Protocol = PROTOCOL_TASKHANDLE;
	reply.Version = PROTOCOL_VERSION;
	reply.UUID = resultuuid;
	reply.Name = taskname;
	
	if (parameters != "") {	
		err = json.Unmarshal ([]byte (parameters), &reply.Parameters);
		if (err != nil) {
			return err;
		}
	}
	
	reply.WorkerSecret = workerSecret;
	return sendJSON (w, &reply);			
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// Updates the status of a task
//////////////////////////////////////////////////////////////////////////////////////////////////////

func TaskUpdateHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, uuid string) error {

	transactionUUID := createUUID ();

	addLogMessage (session, fmt.Sprintf ("Updating task %s (transaction %s)", uuid, transactionUUID), LOGTYPE_TASK_HANDLE, LOGLEVEL_DBONLY);	

	// Parse JSON request
	var request NetTaskUpdateRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_TASKUPDATE);
	if (err != nil) {
		return err;
	}		
	
	workerSecret := request.WorkerSecret;
	
	addLogMessage (session, fmt.Sprintf ("Updating task %s to status %s", uuid, request.Status), LOGTYPE_TASK_HANDLE, LOGLEVEL_CONSOLE);	
	addLogMessage (session, fmt.Sprintf ("  Worker Secret: %s", workerSecret), LOGTYPE_TASK_UPDATE, LOGLEVEL_DBONLY);	
	
	jsondata, err := json.Marshal (&request.Results);
	if (err != nil) {
		return err;
	}
	
	resultjson := string (jsondata);	
	
	addLogMessage (session, fmt.Sprintf ("  Result JSON: %s", resultjson), LOGTYPE_TASK_UPDATE, LOGLEVEL_DBONLY);	
	
	status := request.Status;
	if (status != "SUCCESS") && (status != "ERROR") && (status != "CANCELED") && (status != "RETURNED") {
		return errors.New ("Invalid status string: " + status);
	}


	query := fmt.Sprintf ("UPDATE netstorage_tasks SET status=?, transactionuuid=?, taskresult=? WHERE uuid=? AND status=? AND workersecret=?");
		
	statement1, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}
	
	defer statement1.Close();
	
	_, err = statement1.Exec (status, transactionUUID, resultjson, uuid, "INPROCESS", workerSecret);
	if (err != nil) {
		return err;
	}

	query = fmt.Sprintf ("SELECT transactionuuid FROM netstorage_tasks WHERE uuid=?");
		
	statement2, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}

	defer statement2.Close();
	
	rows, err := statement2.Query (uuid);
	if (err != nil) {
		return err;
	}
	
	defer rows.Close();
	
	if (!rows.Next()) {
		return errors.New ("Could not find job: " + uuid);
	}
	
	foundtransaction := "";	
	err = rows.Scan (&foundtransaction);
	if (err != nil) {
		return err;
	}
	
	if (foundtransaction != transactionUUID) {
		return errors.New ("Could not update job: " + uuid );
	}
			
		
	var reply NetTaskUpdateReply;
	reply.Protocol = PROTOCOL_TASKUPDATE;
	reply.Version = PROTOCOL_VERSION;
	reply.UUID = uuid;
	return sendJSON (w, &reply);			
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// Retrieves the status of a task
//////////////////////////////////////////////////////////////////////////////////////////////////////

func TaskStatusHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, uuid string) error {

	addLogMessage (session, fmt.Sprintf ("Retrieving status of task %s", uuid), LOGTYPE_TASK_STATUS, LOGLEVEL_CONSOLE);	

	// Parse JSON request		
	query := fmt.Sprintf ("SELECT taskname, status, parameters, timestamp, worker, taskresult FROM netstorage_tasks WHERE uuid=?");
		
	statement, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}

	defer statement.Close();
	
	rows, err := statement.Query (uuid);
	if (err != nil) {
		return err;
	}
	
	defer rows.Close();
	
	if (!rows.Next()) {
		return errors.New ("Could not find job: " + uuid);
	}
	
	taskname := "";	
	status := "";
	parameters := "";
	timestampstr := "";
	worker := "";
	taskresult := "";
	err = rows.Scan (&taskname, &status, &parameters, &timestampstr, &worker, &taskresult);
	if (err != nil) {
		return err;
	}
	
	timestamp := timestampstr;		
		
	var reply NetTaskStatusReply;
	reply.Protocol = PROTOCOL_TASKSTATUS;
	reply.Version = PROTOCOL_VERSION;
	reply.UUID = uuid;
	reply.Status = status;
	reply.Name = taskname;
	
	
	if (parameters != "") {	
		err = json.Unmarshal ([]byte (parameters), &reply.Parameters);
		if (err != nil) {
			return err;
		}
	}

	if (taskresult != "") {
		err = json.Unmarshal ([]byte (taskresult), &reply.Result);
		if (err != nil) {
			return err;
		}
	}

	reply.TimeStamp = timestamp;
	reply.Worker = worker;
	
	return sendJSON (w, &reply);			
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// Task handler
//////////////////////////////////////////////////////////////////////////////////////////////////////

func TaskHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) (bool, error) {

	url := r.URL.Path;
	uuid := "";

	if (r.Method == "POST") {		
		
		if urlCheckRootURL (url, "tasks/new", true) {
			err := TaskNewHandler (db, session, w, r);
			return true, err;
		}

		if urlCheckRootURL (url, "tasks/clear", true) {
			err := TaskClearHandler (db, session, w, r);
			return true, err;
		}

		if urlCheckRootURL (url, "tasks/handle", true) {
			err := TaskHandleHandler (db, session, w, r);
			return true, err;
		}

		if parseUUIDURL (url, "tasks", "", &uuid) {
			err := TaskUpdateHandler (db, session, w, r, uuid);
			return true, err;
		}		
		
		
	}
	

	if (r.Method == "GET") {
		if parseUUIDURL (url, "tasks", "", &uuid) {
			err := TaskStatusHandler (db, session, w, r, uuid);
			return true, err;
		}		
	}
	
	return false, nil;
	
}


