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
// netfabbstorage_auth.go
// Handles a fast user authentication mechanism
//////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"time"
)

type LogLevel int
type LogType string

const (
	LOGLEVEL_CONSOLE = 1
	LOGLEVEL_DBONLY  = 2
	LOGLEVEL_DEBUG   = 3
)

const (
	LOGTYPE_SYSTEM     = "SYSTEM"
	LOGTYPE_ORM_UPDATE = "ORMUPD"
	LOGTYPE_ORM_READ   = "ORMRED"
	LOGTYPE_ORM_DELETE = "ORMDEL"
	LOGTYPE_ORM_SAVE   = "ORMSAV"

	LOGTYPE_TASK_CLEAR  = "TSKCLR"
	LOGTYPE_TASK_UPDATE = "TSKUPD"
	LOGTYPE_TASK_HANDLE = "TSKHND"
	LOGTYPE_TASK_STATUS = "TSKSTA"
	LOGTYPE_TASK_NEW    = "TSKNEW"

	LOGTYPE_DATA_HUBS           = "DATHUB"
	LOGTYPE_DATA_PROJECTS       = "DATPRJ"
	LOGTYPE_DATA_ROOTFOLDERS    = "DATRFL"
	LOGTYPE_DATA_SUBFOLDERS     = "DATSFL"
	LOGTYPE_DATA_ITEMS          = "DATITM"
	LOGTYPE_DATA_ENTITIES       = "DATENT"
	LOGTYPE_DATA_NEWPROJECT     = "DATNPR"
	LOGTYPE_DATA_NEWFOLDER      = "DATNFL"
	LOGTYPE_DATA_NEWITEM        = "DATNIT"
	LOGTYPE_DATA_UPLOADBINARY   = "DATUPL"
	LOGTYPE_DATA_UPLOADENTITY   = "DATUEN"
	LOGTYPE_DATA_DOWNLOADENTITY = "DATDEN"

	LOGTYPE_PANSERVICE = "SVCPAN"
)

var SessionDBFileName string
var SessionDB *sql.DB = nil

//////////////////////////////////////////////////////////////////////////////////////////////////////
// initSessionDB
// Initializes the session and log database
//////////////////////////////////////////////////////////////////////////////////////////////////////

func initSessionDB(dbPrefix string) (error, string) {
	currentTime := time.Now()

	SessionDBFileName := dbPrefix + currentTime.Format("20060102_150405") + ".db"

	db, err := sql.Open("sqlite3", SessionDBFileName)
	if err != nil {
		return err, SessionDBFileName
	}
	db.SetMaxOpenConns(1)

	query := "CREATE TABLE `sessions` (" +
		"`sessionuuid`	varchar ( 64 ) NOT NULL UNIQUE," +
		"`token`	varchar ( 512 ) NOT NULL, " +
		"`userid`	varchar ( 64 ) NOT NULL, " +
		"`status`	varchar (32) NOT NULL," +
		"`timestamp` varchar ( 64 ) NOT NULL" +
		")"

	_, err = db.Exec(query)
	if err != nil {
		return err, SessionDBFileName
	}

	query = "CREATE TABLE `logs` (" +
		"`loguuid`	varchar ( 64 ) NOT NULL, " +
		"`logindex`	int DEFAULT 0, " +
		"`sessionuuid`	varchar ( 64 ) NOT NULL," +
		"`userid`	varchar ( 64 ) NOT NULL," +
		"`logtype`	varchar ( 6 ) NOT NULL," +
		"`timestamp` varchar ( 64 ) NOT NULL, " +
		"`message`	TEXT DEFAULT 1 )"

	_, err = db.Exec(query)

	SessionDB = db

	return err, SessionDBFileName
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// createEmptySession
// creates an empty session
//////////////////////////////////////////////////////////////////////////////////////////////////////

func createEmptySession() NetStorageSession {
	var session NetStorageSession
	session.LogUUID = createUUID()
	session.LogIndex = 1

	return session
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// createNewSessionInDB
// creates a session and a token in the DB
//////////////////////////////////////////////////////////////////////////////////////////////////////

func createNewSessionInDB(UserID string) (NetStorageSession, error) {

	session := createEmptySession()
	session.UUID = ""
	session.UserID = ""
	session.Token = ""
	session.Active = 0

	if !userIsValidIdentifier(UserID) {
		return session, errors.New("userID contained invalid identifier!")
	}

	newUUID := createUUID()

	var tokenData NetStorageToken
	tokenData.SessionUUID = newUUID
	tokenData.UserID = UserID

	tokenJSONData, err := json.Marshal(&tokenData)
	if err != nil {
		return session, err
	}

	newToken := base64.StdEncoding.EncodeToString(tokenJSONData)

	if len(UserID) == 0 {
		return session, errors.New("Invalid UserID for session!")
	}

	timestamp := time.Now().Format(time.RFC3339)

	statement, err := SessionDB.Prepare("INSERT INTO sessions (sessionuuid, userid, token, status, timestamp) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return session, err
	}

	_, err = statement.Exec(newUUID, UserID, newToken, "NEW", timestamp)
	if err != nil {
		return session, err
	}

	session.UUID = newUUID
	session.UserID = UserID
	session.Token = newToken
	session.Active = 0

	return session, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// createNewSessionInDB
// creates a session and a token in the DB
//////////////////////////////////////////////////////////////////////////////////////////////////////

func AuthenticateSessionInDB(SessionUUID string, AuthType string, AuthKey string) (string, error) {

	formattedUUID, err := checkUUIDFormat(SessionUUID)
	if err != nil {
		return "", err
	}

	if AuthType != "saltedhash" {
		return "", errors.New("Invalid Authentication type!")
	}

	AuthenticationVerified := true

	if !AuthenticationVerified {
		return "", errors.New("Could not verify authentication key!")
	}

	statement1, err := SessionDB.Prepare("UPDATE sessions SET status=? WHERE sessionuuid=? AND status=?")
	if err != nil {
		return "", err
	}

	_, err = statement1.Exec("ACCEPTED", formattedUUID, "NEW")
	if err != nil {
		return "", err
	}

	statement2, err := SessionDB.Prepare("SELECT token FROM sessions WHERE sessionuuid=? AND status=?")
	if err != nil {
		return "", err
	}

	rows, err := statement2.Query(formattedUUID, "ACCEPTED")
	if err != nil {
		return "", err
	}

	defer rows.Close()

	if !rows.Next() {
		return "", errors.New("Invalid session uuid: " + formattedUUID)
	}

	Token := ""
	err = rows.Scan(&Token)
	if err != nil {
		return "", err
	}

	return Token, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveUserIDFromSession
// retrieves the userID of a given session.
//////////////////////////////////////////////////////////////////////////////////////////////////////
func RetrieveUserIDFromSession(sessionUUID string) (string, error) {

	statement, err := SessionDB.Prepare("SELECT userid FROM sessions WHERE sessionuuid=?")
	if err != nil {
		return "", err
	}

	rows, err := statement.Query(sessionUUID)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	if !rows.Next() {
		return "", errors.New("Invalid session UUID: " + sessionUUID)
	}

	userID := ""

	err = rows.Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil

}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveSessionByToken
// retrieves a session given by the token
//////////////////////////////////////////////////////////////////////////////////////////////////////

func retrieveSessionByToken(token string, maxSessionDuration int) (NetStorageSession, error) {
	session := createEmptySession()

	session.UUID = ""
	session.UserID = ""
	session.Token = ""
	session.Active = 0

	statement, err := SessionDB.Prepare("SELECT sessionuuid, userid, timestamp FROM sessions WHERE token=? AND status=?")
	if err != nil {
		return session, err
	}

	rows, err := statement.Query(token, "ACCEPTED")
	if err != nil {
		return session, err
	}

	defer rows.Close()

	if !rows.Next() {
		return session, errors.New("Invalid session token: " + token)
	}

	timestamp := ""
	sessionUUID := ""
	sessionUserID := ""

	err = rows.Scan(&sessionUUID, &sessionUserID, &timestamp)
	if err != nil {
		return session, err
	}

	creationTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return session, err
	}

	duration := time.Since(creationTime)

	if duration < 0 {
		return session, errors.New("Invalid session time!")
	}

	if duration.Seconds() >= float64(maxSessionDuration) {
		return session, errors.New("Session expired!")
	}

	session.Active = 1
	session.Token = token
	session.UUID = sessionUUID
	session.UserID = sessionUserID

	return session, nil
}

func addLogMessage(session *NetStorageSession, message string, logtype LogType, loglevel LogLevel) {

	if loglevel == LOGLEVEL_CONSOLE {
		log.Println(logtype, " - ", message)
	}

	if (loglevel == LOGLEVEL_CONSOLE) || (loglevel == LOGLEVEL_DBONLY) {

		loguuid := session.LogUUID

		// Increment log index
		session.Mutex.Lock()
		logindex := session.LogIndex
		session.LogIndex = logindex + 1
		session.Mutex.Unlock()

		timestamp := time.Now().Format(time.RFC3339)

		if SessionDB == nil {
			log.Fatal("could not log message to database!")
		}

		statement, err := SessionDB.Prepare("INSERT INTO logs (loguuid, logindex, sessionuuid, userid, logtype, timestamp, message) VALUES(?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal("could not log prepare logging: ", err)
		}

		_, err = statement.Exec(loguuid, logindex, session.UUID, session.UserID, logtype, timestamp, message)
		if err != nil {
			log.Fatal("could not execute logging: ", err)
		}

	}

}
