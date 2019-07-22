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
// netfabbstorage_orm.go
// Handles netfabb Object Relational Model requests.
//////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"net/http"
	"encoding/base64"
	"database/sql"	
	"io/ioutil"
	"os"
	"regexp"
	"encoding/json"
	"errors"
)


var ORMSchema NetORMSchema;
var ORMSchemaTable map[string]NetORMTableMapping;

var ORMInputCheck = regexp.MustCompile(`^[A-Z_]+$`).MatchString


//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORMInitialiseSchema
// Loads ORM Schema from JSON file
//////////////////////////////////////////////////////////////////////////////////////////////////////
func ORMInitialiseSchema (filename string) (int, error) {

	file, err := os.Open(filename);
	if (err != nil) {
		return 0, err;
	}
	
	defer file.Close();
	
	var schema NetORMSchema; 
	jsondata, err := ioutil.ReadAll (file);
	if (err != nil) {
		return 0, err;
	}

	err = json.Unmarshal (jsondata, &schema);
	if (err != nil) {
		return 0, err;
	}
	
	if (schema.Schema != PROTOCOL_ORMSCHEMA) {
		return 0, errors.New ("Invalid ORM Schema Type: " + schema.Schema);
	}
	if (schema.Version != PROTOCOL_VERSION) {
		return 0, errors.New ("Invalid ORM Schema Version: " + schema.Version);
	}
	
	count := 0;

	ORMSchemaTable = make(map[string]NetORMTableMapping);
	for _, mapping := range schema.Mappings {
	
		if (!ORMInputCheck (mapping.Name)) {
			return 0, errors.New ("Invalid Mapping Name: " + mapping.Name);
		}
		
		var tablemapping NetORMTableMapping;
		tablemapping.Name = mapping.Name;
		tablemapping.PrimaryKey = mapping.PrimaryKey;
		tablemapping.FieldMap = make(map[string]NetORMValue);
		
		for _, field := range mapping.Fields {
			if (!ORMInputCheck (field.Key)) {
				return 0, errors.New ("Invalid Mapping Key: " + field.Key);
			}
			tablemapping.FieldMap[field.Key] = field;
		}
		
	
		ORMSchemaTable[tablemapping.Name] = tablemapping;
		
		count = count + 1;
	}
	
	return count, nil;
		
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORMMappingExists
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMMappingExists (name string) bool {
	if (ORMSchemaTable != nil) {
		_, mappingexists := ORMSchemaTable[name];
		return mappingexists;
	}
	
	return false;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORMKeyExists
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMKeyExists (mappingname string, key string) bool {
	if (ORMSchemaTable != nil) {
		mapping, mappingexists := ORMSchemaTable[mappingname];
		if (mappingexists) {
			_, keyexists := mapping.FieldMap[key];
			return keyexists;
		}
	}
	
	return false;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORMKeyExists
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMGetKeyType (mappingname string, key string) string {
	if (ORMSchemaTable != nil) {
		mapping, mappingexists := ORMSchemaTable[mappingname];
		if (mappingexists) {
			key, keyexists := mapping.FieldMap[key];
			if (keyexists) {
				return key.Type;
			}
		}
	}
	
	return "";
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORMGetKeys
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMGetMapping (mappingname string) NetORMTableMapping {
	
	var mapping NetORMTableMapping;
	
	if (ORMSchemaTable != nil) {
		mapping, _ = ORMSchemaTable[mappingname];	
	}
	
	return mapping;
}



//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORMMapValueToByteField
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMMapValueToByteField (value * NetORMValue, intf * interface{}) error {
	if (value.Type == "integer") || (value.Type == "varchar") || (value.Type == "boolean") || (value.Type == "datetime") || (value.Type == "uuid") {
		*intf = value.Value;
	}
	
	if (value.Type == "blob") {
		decodedbytes, err := base64.StdEncoding.DecodeString (value.Value);
		if (err != nil) {
			return err;
		}
		
		*intf = decodedbytes;
	}

	return nil;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORMMapValueToByteField
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMMapByteFieldToValue (valuetype string, value *sql.RawBytes) (string, error) {
	if (valuetype == "integer") || (valuetype == "varchar") || (valuetype == "boolean") || (valuetype == "datetime") || (valuetype == "uuid") {
		return string(*value), nil;
	}
	
	if (valuetype == "blob") {
		encodedstring := base64.StdEncoding.EncodeToString (*value);		
		return encodedstring, nil;
	}

	return "", errors.New("Invalid value type");
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORM read handler
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMReadHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) error {
	addLogMessage (session, "ORM read request.", LOGTYPE_ORM_READ, LOGLEVEL_CONSOLE);	
	
	// Parse JSON request
	var request NetORMReadRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_ORMREAD);
	if (err != nil) {
		return err;
	}		
	
	entity := request.Entity;
	if (!ORMInputCheck (entity)) {
		return errors.New ("Invalid Mapping Name: " + entity);
	}
	
	if (!ORMMappingExists (entity)) {
		return errors.New ("invalid ORM entity: " + entity);		
	}
	
	addLogMessage (session, "read request entity: " + entity, LOGTYPE_ORM_READ, LOGLEVEL_CONSOLE);	
		
	query := "SELECT";
	isFirst := true;
	for _, value := range request.Values {
		if (isFirst) {
			isFirst = false;
		} else {
			query = query + ",";
		}
		
		if (!ORMInputCheck (value.Key)) {
			return errors.New ("Invalid Key Name: " + value.Key);
		}
		if (!ORMKeyExists (entity, value.Key)) {
			return errors.New ("invalid ORM key: " + value.Key);
		}
		
		query = query + " " + value.Key;
	}

	query = query + " FROM " + entity + " WHERE SYS_ACTIVE=1";
	
	filterarray := make([]interface{}, len(request.Filter));
	index := 0;
	
	for _, value := range request.Filter {
		query = query + " AND ";

		if (!ORMInputCheck (value.Key)) {
			return errors.New ("Invalid Key Name: " + value.Key);
		}
		if (!ORMKeyExists (entity, value.Key)) {
			return errors.New ("invalid ORM key: " + value.Key);
		}
		
		ORMMapValueToByteField (&value, &filterarray[index]);
		index = index + 1;
		
		query = query + value.Key + "=? ";
	}
		
	statement, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}
	
	rows, err := statement.Query (filterarray...);

	if (err != nil) {
		return err;
	}
	
	defer rows.Close();
	
	columnNames, err := rows.Columns();
	if (err != nil) {
		return err;
	}
	defer rows.Close ();
	
	lenCN := len(columnNames);	
	
	resultrows := make([] NetORMRow, 0);	
	
	for (rows.Next()) {
	
		columnPointers := make([]interface{}, lenCN)
		for i := 0; i < lenCN; i++ {
			columnPointers[i] = new(sql.RawBytes)
		}	
				
		err = rows.Scan(columnPointers...);
		if (err != nil) {
			return err;
		}
		
		var entry NetORMRow;
		entry = make([] string, lenCN);

		for i := 0; i < lenCN; i++ {
			rawbytes, _ := columnPointers[i].(*sql.RawBytes);
			keytype := ORMGetKeyType (entity, columnNames[i]);
			
			entry[i], err = ORMMapByteFieldToValue (keytype, rawbytes);
			if (err != nil) {
				return err;
			}
			
		}	
		
		resultrows = append (resultrows, entry);	
	}	
	
	
	var reply NetORMReadReply;
	reply.Protocol = PROTOCOL_ORMREAD;
	reply.Version = PROTOCOL_VERSION;
	reply.Rows = resultrows;
	reply.Columns = make([] string, lenCN);
	for i := 0; i < lenCN; i++ {
		reply.Columns[i] = columnNames[i];
	}	
	
	return sendJSON (w, &reply);			
}



//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORM delete handler
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMDeleteHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) error {
	// Parse JSON request
	var request NetORMDeleteRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_ORMDELETE);
	if (err != nil) {
		return err;
	}		
	
	entity := request.Entity;
	if (!ORMInputCheck (entity)) {
		return errors.New ("Invalid Mapping Name: " + entity);
	}
	if (!ORMMappingExists (entity)) {
		return errors.New ("invalid ORM entity: " + entity);		
	}
	
	addLogMessage (session, "delete request entity: " + entity, LOGTYPE_ORM_DELETE, LOGLEVEL_CONSOLE);	
		
	query := "UPDATE " + entity + " SET SYS_ACTIVE=0 WHERE SYS_ACTIVE=1";
	
	filterarray := make([]interface{}, len(request.Filter));
	index := 0;
	
	for _, value := range request.Filter {
		query = query + " AND ";
		
		if (!ORMInputCheck (value.Key)) {
			return errors.New ("Invalid Key Name: " + value.Key);
		}
		if (!ORMKeyExists (entity, value.Key)) {
			return errors.New ("invalid ORM key: " + value.Key);
		}
		
		ORMMapValueToByteField (&value, &filterarray[index]);
		
		query = query + value.Key + "=?";
		index = index + 1;
	}

	statement, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}
	
	_, err = statement.Exec (filterarray...);
	if (err != nil) {
		return err;
	}
		
	var reply NetORMDeleteReply;
	reply.Protocol = PROTOCOL_ORMDELETE;
	reply.Version = PROTOCOL_VERSION;
	return sendJSON (w, &reply);			
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORM save handler
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMSaveHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) error {
	addLogMessage (session, "ORM save request.", LOGTYPE_ORM_SAVE, LOGLEVEL_CONSOLE);	
	
	// Parse JSON request
	var request NetORMSaveRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_ORMSAVE);
	if (err != nil) {
		return err;
	}		
	
	entity := request.Entity;
	if (!ORMInputCheck (entity)) {
		return errors.New ("Invalid Mapping Name: " + entity);
	}	
	if (!ORMMappingExists (entity)) {
		return errors.New ("invalid ORM entity: " + entity);		
	}

	addLogMessage (session, "save request entity: " + entity, LOGTYPE_ORM_SAVE, LOGLEVEL_CONSOLE);	
	
	query := "INSERT INTO " + entity + "(";
	isFirst := true;
	for _, value := range request.Values {
		if (isFirst) {
			isFirst = false;
		} else {
			query = query + ",";
		}
		
		if (!ORMInputCheck (value.Key)) {
			return errors.New ("Invalid Key Name: " + value.Key);
		}
		if (!ORMKeyExists (entity, value.Key)) {
			return errors.New ("invalid ORM key: " + value.Key);
		}
		
		query = query + " " + value.Key;
	}
	
	query = query + ") VALUES (";

	valuearray := make([]interface{}, len(request.Values));
	index := 0;
	for _, value := range request.Values {
		if (index != 0) {
			query = query + ",";
		}
		
		ORMMapValueToByteField (&value, &valuearray[index]);
		query = query + " ?";
		index = index + 1;
	}
	query = query + ")";
	
	statement, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}
	
	_, err = statement.Exec (valuearray...);
	if (err != nil) {
		return err;
	}
	
	
	var reply NetORMSaveReply;
	reply.Protocol = PROTOCOL_ORMSAVE;
	reply.Version = PROTOCOL_VERSION;
	return sendJSON (w, &reply);			
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORM update handler
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMcopyToArchive (db *sql.DB, request NetORMUpdateRequest) error {

	entity := request.Entity;

	mapping := ORMGetMapping (entity);
	
	
	fieldliststring := "SYS_ACTIVE, SYS_VERSION";
	for _, value := range mapping.FieldMap {				
		fieldliststring = fieldliststring + ", " + value.Key;
	}
		
		
	query := "INSERT INTO " + entity + "_ARCHIVE (" + fieldliststring;
	query = query + ") SELECT " + fieldliststring;	
	query = query + " FROM " + entity + " WHERE SYS_ACTIVE=1";
	
	filterarray := make([]interface{}, len(request.Filter));
	index := 0;
	
	for _, value := range request.Filter {
		query = query + " AND ";

		if (!ORMInputCheck (value.Key)) {
			return errors.New ("Invalid Key Name: " + value.Key);
		}
		if (!ORMKeyExists (entity, value.Key)) {
			return errors.New ("invalid ORM key: " + value.Key);
		}
		
		ORMMapValueToByteField (&value, &filterarray[index]);
		index = index + 1;
		
		query = query + value.Key + "=? ";
	}
		
	statement, err := db.Prepare (query);
	if (err != nil) {
		return err;
	}
	
	_, err = statement.Exec (filterarray...);
	
	return err;
}


func ORMUpdateHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) error {
	addLogMessage (session, "ORM update request.", LOGTYPE_ORM_UPDATE, LOGLEVEL_CONSOLE);	

	// Parse JSON request
	var request NetORMUpdateRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_ORMUPDATE);
	if (err != nil) {
		return err;
	}		
	
	entity := request.Entity;
	if (!ORMInputCheck (entity)) {
		return errors.New ("Invalid Mapping Name: " + entity);
	}	
	if (!ORMMappingExists (entity)) {
		return errors.New ("invalid ORM entity: " + entity);		
	}

	addLogMessage (session, "update request entity: " + entity, LOGTYPE_ORM_UPDATE, LOGLEVEL_CONSOLE);	
	
	
	err = BeginTransaction (db)
	if (err != nil) {
		return err;
	}		


	err = ORMcopyToArchive (db, request);
	if (err != nil) {
		RollbackTransaction (db);
		return err;
	}		
	
		
	query := "UPDATE " + entity + " SET SYS_VERSION=SYS_VERSION + 1";
	filterarray := make([]interface{}, len(request.Filter) + len(request.Values));
	index := 0;
	
	for _, value := range request.Values {
		query = query + ",";
		
		if (!ORMInputCheck (value.Key)) {
			RollbackTransaction (db);
			return errors.New ("Invalid Key Name: " + value.Key);
		}
		if (!ORMKeyExists (entity, value.Key)) {
			RollbackTransaction (db);
			return errors.New ("invalid ORM key: " + value.Key);
		}
		
		ORMMapValueToByteField (&value, &filterarray[index]);
		query = query + " " + value.Key + "=?";
		index = index + 1;
	}

	query = query + " WHERE SYS_ACTIVE=1";
		
	for _, value := range request.Filter {
		query = query + " AND ";

		if (!ORMInputCheck (value.Key)) {
			RollbackTransaction (db);
			return errors.New ("Invalid Key Name: " + value.Key);
		}
		if (!ORMKeyExists (entity, value.Key)) {
			RollbackTransaction (db);
			return errors.New ("invalid ORM key: " + value.Key);
		}
		
		ORMMapValueToByteField (&value, &filterarray[index]);
		index = index + 1;
		
		query = query + value.Key + "=? ";
	}
	
		
	statement, err := db.Prepare (query);
	if (err != nil) {
		RollbackTransaction (db);
		return err;
	}
	
	_, err = statement.Exec (filterarray...);

	if (err != nil) {
		RollbackTransaction (db);
		return err;
	}
	
	err = CommittTransaction (db);
	if (err != nil) {
		return err;
	}
	
	
	var reply NetORMUpdateReply;
	reply.Protocol = PROTOCOL_ORMUPDATE;
	reply.Version = PROTOCOL_VERSION;
	return sendJSON (w, &reply);			
}



//////////////////////////////////////////////////////////////////////////////////////////////////////
// ORM handler
//////////////////////////////////////////////////////////////////////////////////////////////////////

func ORMHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) (bool, error) {

	url := r.URL.Path;

	if (r.Method == "POST") {		
		
		if urlCheckRootURL (url, "orm/read", true) {
			err := ORMReadHandler (db, session, w, r);
			return true, err;
		}

		if urlCheckRootURL (url, "orm/save", true) {
			err := ORMSaveHandler (db, session, w, r);
			return true, err;
		}
		
		if urlCheckRootURL (url, "orm/delete", true) {
			err := ORMDeleteHandler (db, session, w, r);
			return true, err;
		}

		if urlCheckRootURL (url, "orm/update", true) {
			err := ORMUpdateHandler (db, session, w, r);
			return true, err;
		}
	
	}
	
	return false, nil;
	
}


