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
// netfabbstorage_projects.go
// Handles netfabb Project Data requests.
//////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"net/http"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"database/sql"	
	"crypto/sha1"
)



//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageHubsHandler
// handles a /hubs GET request and retrieves a list of hubs.
//////////////////////////////////////////////////////////////////////////////////////////////////////


func StorageHubsHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) error {
	addLogMessage (session, "Retrieving Hubs", LOGTYPE_DATA_HUBS, LOGLEVEL_CONSOLE);

	hubs, err := RetrieveHubs (db);
	if (err == nil) {
		var reply NetStorageHubReply;
		reply.Protocol = PROTOCOL_HUBS;
		reply.Version = PROTOCOL_VERSION;
		reply.Hubs = hubs;			
		return sendJSON (w, &reply);			
	}			
	
	return err;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageProjectsHandler
// handles a /hubs/<uuid> GET request and returns all projects for a given hub
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageProjectsHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, hubuuid string) error {
	addLogMessage (session, "Retrieving Projects for Hub: " + hubuuid, LOGTYPE_DATA_PROJECTS, LOGLEVEL_CONSOLE);	

	projects, err := RetrieveProjects (db, hubuuid);
	if (err == nil) {	
		var reply NetStorageProjectsReply;
		reply.Protocol = PROTOCOL_PROJECTS;
		reply.Version = PROTOCOL_VERSION;
		reply.HubUUID = hubuuid;
		reply.Projects = projects;	
		return sendJSON (w, &reply);			
	}			
	
	return err;
}





//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageRootFoldersHandler
// handles a /projects/<uuid>/rootfolders GET request and retrieves a list of all folders for a project
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageRootFoldersHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, projectuuid string) error {
	addLogMessage (session, "Retrieving Folders for Project: " + projectuuid, LOGTYPE_DATA_ROOTFOLDERS, LOGLEVEL_CONSOLE);

	folders, err := RetrieveRootFoldersOfProject (db, projectuuid);
	if (err == nil) {	
		var reply NetStorageFoldersReply;
		reply.Protocol = PROTOCOL_ROOTFOLDERS;
		reply.Version = PROTOCOL_VERSION;
		reply.Folders = folders;	
		return sendJSON (w, &reply);	
	}			
	
	return err;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageSubFoldersHandler
// handles a /folders/<uuid>/subfolders GET request and retrieves a list of all subfolders of a folder
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageSubFoldersHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, folderuuid string) error {
	addLogMessage (session, "Retrieving Folders for Folder: " + folderuuid, LOGTYPE_DATA_SUBFOLDERS, LOGLEVEL_CONSOLE);

	folders, err := RetrieveSubFoldersOfFolder (db, folderuuid);
	if (err == nil) {	
		var reply NetStorageFoldersReply;
		reply.Protocol = PROTOCOL_SUBFOLDERS;
		reply.Version = PROTOCOL_VERSION;
		reply.Folders = folders;	
		return sendJSON (w, &reply);			
	}			
	
	return err;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageSubItemsHandler
// handles a /folders/<uuid>/items GET request and retrieves a list of all items of a folder
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageItemsHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, folderuuid string) error {
	addLogMessage (session, "Retrieving Items for Folder: " + folderuuid, LOGTYPE_DATA_ITEMS, LOGLEVEL_CONSOLE);	

	items, err := RetrieveItemsOfFolder (db, folderuuid);
	if (err == nil) {	
		var reply NetStorageItemsReply;
		reply.Protocol = PROTOCOL_ITEMS;
		reply.Version = PROTOCOL_VERSION;
		reply.Items = items;		
		return sendJSON (w, &reply);			
	}			
	
	return err;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageEntitiesHandler
// handles a /items/<uuid>/entities GET request and retrieves a list of all items of a folder
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageEntitiesHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, itemuuid string) error {
	addLogMessage (session, "Retrieving Entities for Items: " + itemuuid, LOGTYPE_DATA_ENTITIES, LOGLEVEL_CONSOLE);	
 
	entities, err := RetrieveEntitiesOfItem (db, itemuuid);
	if (err == nil) {	
		var reply NetStorageEntitiesReply;
		reply.Protocol = PROTOCOL_ENTITIES;
		reply.Version = PROTOCOL_VERSION;
		reply.Entities = entities;		
		return sendJSON (w, &reply);			
	}			
	
	return err;
}



//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageProjectNewHandler
// handles a /hubs/<uuid> POST request and creates a new project. It returns the uuid of the created 
// project.
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageProjectNewHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, hubuuid string) error {
	addLogMessage (session, "Creating new project for hub: " + hubuuid, LOGTYPE_DATA_NEWPROJECT, LOGLEVEL_CONSOLE);	
	
	// Parse JSON request
	var request NetStorageNewProjectRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_NEWPROJECT);
	if (err != nil) {
		return err;
	}
			
	// create new Folder
	projectuuid := createUUID ();
	rootfolderuuid := createUUID ();
	
	err = BeginTransaction (db);
	if (err != nil) {
		return err;
	}

	// create project entry
	err = createNewProject (db, projectuuid, request.ProjectName, hubuuid);
	if (err != nil) {
		RollbackTransaction (db);
		return err;
	}
	
	// create folder entry
	err = createNewFolder (db, rootfolderuuid, projectuuid, request.ProjectName, "");
	if (err != nil) {
		RollbackTransaction (db);
		return err;
	}
	
	err = CommittTransaction (db);
	if (err != nil) {
		return err;
	}

		
	// Send reply JSON	
	var reply NetStorageNewProjectReply;
	reply.Protocol = PROTOCOL_NEWPROJECT;
	reply.Version = PROTOCOL_VERSION;
	reply.HubUUID = hubuuid;
	reply.ProjectUUID = projectuuid;	
	reply.RootFolderUUID = rootfolderuuid;
	return sendJSON (w, &reply);			
	

	
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageFolderNewHandler
// handles a /folders/<uuid>/newfolder POST request and creates a new subfolder. It returns the uuid 
// of the created subfolder.
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageFolderNewHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, parentuuid string) error {
	addLogMessage (session, "Creating new subfolder for folder: " + parentuuid, LOGTYPE_DATA_NEWFOLDER, LOGLEVEL_CONSOLE);	
	
	// Parse JSON request
	var request NetStorageNewFolderRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_NEWFOLDER);
	if (err != nil) {
		return err;
	}
	
	folder, err := RetrieveFolderByUUID (db, parentuuid);
	if (err != nil) {
		return err;
	}
		
	// create new Folder
	subfolderuuid := createUUID ();
	
	err = BeginTransaction (db);
	if (err != nil) {
		return err;
	}

	// create folder entry
	err = createNewFolder (db, subfolderuuid, folder.ProjectUUID, request.FolderName, folder.UUID);
	if (err != nil) {
		RollbackTransaction (db);
		return err;
	}
	
	err = CommittTransaction (db);
	if (err != nil) {
		return err;
	}

		
	// Send reply JSON	
	var reply NetStorageNewFolderReply;
	reply.Protocol = PROTOCOL_NEWFOLDER;
	reply.Version = PROTOCOL_VERSION;
	reply.ProjectUUID = folder.ProjectUUID;	
	reply.ParentUUID = folder.UUID;
	reply.SubFolderUUID = subfolderuuid;
	return sendJSON (w, &reply);			
		
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageItemNewHandler
// handles a /folders/<uuid>/newitem POST request and creates a new item. It returns the uuid 
// of the created item.
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageItemNewHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, folderuuid string) error {
	addLogMessage (session, "Creating new item for folder: " + folderuuid, LOGTYPE_DATA_NEWITEM, LOGLEVEL_CONSOLE);	
	
	// Parse JSON request
	var request NetStorageNewItemRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_NEWITEM);
	if (err != nil) {
		return err;
	}
	
	folder, err := RetrieveFolderByUUID (db, folderuuid);
	if (err != nil) {
		return err;
	}
		
	// create new Item
	itemuuid := createUUID ();
	
	err = BeginTransaction (db);
	if (err != nil) {
		return err;
	}

	// create folder entry
	err = createNewItem (db, itemuuid, folder.UUID, request.ItemName);
	if (err != nil) {
		RollbackTransaction (db);
		return err;
	}
	
	err = CommittTransaction (db);
	if (err != nil) {
		return err;
	}

		
	// Send reply JSON	
	var reply NetStorageNewItemReply;
	reply.Protocol = PROTOCOL_NEWITEM;
	reply.Version = PROTOCOL_VERSION;
	reply.ItemUUID = itemuuid;	
	reply.FolderUUID = folderuuid;
	return sendJSON (w, &reply);			
	
	
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageBinaryUploadHandler
// handles a /upload/<uuid> POST request and creates a new entity for the item.
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageBinaryUploadHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, itemuuid string) error {
	addLogMessage (session, "Uploading data for item: " + itemuuid, LOGTYPE_DATA_UPLOADBINARY, LOGLEVEL_CONSOLE);	
			
	// retrieve Item by UUID
	item, err := RetrieveItemByUUID (db, itemuuid);
	if (err != nil) {
		return err;
	}
	
	entityuuid := createUUID ();	
	
	// calculate sha1 sum
	bytes, err := ioutil.ReadAll (r.Body);
	if (err != nil) {
		return err;
	}			
	sha1sum := fmt.Sprintf("%x", sha1.Sum(bytes));		
	
	filesize := int64 (len(bytes));
		
	// create new Entity
	err = createNewEntity (db, entityuuid, item.UUID, sha1sum, filesize, false);
	if (err != nil) {
		return err;
	}	
	
	// create file on disk
	file, err := os.Create(getUUIDStorageName (entityuuid));
	if (err != nil) {
		return err;
	}	
	defer file.Close();
		
	// copy data
	_, err = file.Write(bytes);
	if err != nil {
		return err;
	}
		
	// Send reply JSON	
	var reply NetStorageNewEntityReply;
	reply.Protocol = PROTOCOL_NEWENTITY;
	reply.Version = PROTOCOL_VERSION;
	reply.ItemUUID = itemuuid;	
	reply.EntityUUID = entityuuid;
	return sendJSON (w, &reply);			
	
	return err;
	
	
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageEntityUpdateHandler
// updates an entity db entry
//////////////////////////////////////////////////////////////////////////////////////////////////////

func StorageEntityUpdateHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, entityuuid string) error {
	addLogMessage (session, "Updating entity: " + entityuuid, LOGTYPE_DATA_UPLOADENTITY, LOGLEVEL_CONSOLE);	
	
	// Parse JSON request
	var request NetStorageUpdateEntityRequest;
	err := parseJSONRequest (r, &request, PROTOCOL_UPDATEENTITY);
	if (err != nil) {
		return err;
	}
	
	err = BeginTransaction (db);
	if (err != nil) {
		return err;
	}
		
	entity, err := RetrieveEntityByUUID (db, entityuuid, false);
	if (err != nil) {
		RollbackTransaction (db);
		return err;
	}
		
	err = updateEntity (db, entity.UUID, request.DataType, string (request.MetaData), true);
	if (err != nil) {
		RollbackTransaction (db);
		return err;
	}

	err = CommittTransaction (db);
	if (err != nil) {
		return err;
	}
	

	// Send reply JSON	
	var reply NetStorageUpdateEntityReply;
	reply.Protocol = PROTOCOL_UPDATEENTITY;
	reply.Version = PROTOCOL_VERSION;
	reply.ItemUUID = entity.ItemUUID;	
	reply.EntityUUID = entity.UUID;
	return sendJSON (w, &reply);			
	
	
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// StorageDownloadHandler
// downloads the content of an entity
//////////////////////////////////////////////////////////////////////////////////////////////////////


func StorageDownloadHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request, entityuuid string) error {
	addLogMessage (session, "Downloading entity: " + entityuuid, LOGTYPE_DATA_DOWNLOADENTITY, LOGLEVEL_CONSOLE);	
			
	entity, err := RetrieveEntityByUUID (db, entityuuid, false);
	if (err != nil) {
		return err;
	}
		
	file, err := os.Open(getUUIDStorageName (entity.UUID));
	if (err != nil) {
		return err;
	}
	
	defer file.Close();

	_, err = io.Copy(w, file)
	return err;
	
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// Data handler
//////////////////////////////////////////////////////////////////////////////////////////////////////

func DataHandler (db *sql.DB, session * NetStorageSession, w http.ResponseWriter, r *http.Request) (bool, error) {

	url := r.URL.Path;
	uuid := "";

	if (r.Method == "GET") {		
	
		if urlCheckRootURL (url, "data/hubs", true) {
			err := StorageHubsHandler (db, session, w, r);
			return true, err;
		}
		
		if parseUUIDURL (url, "data/hubs", "", &uuid) {
			err := StorageProjectsHandler (db, session, w, r, uuid);
			return true, err;
		}

		if parseUUIDURL (url, "data/projects", "rootfolders", &uuid) {
			err := StorageRootFoldersHandler (db, session, w, r, uuid);
			return true, err;
		}
		
		if parseUUIDURL (url, "data/folders", "subfolders", &uuid) {
			err := StorageSubFoldersHandler (db, session, w, r, uuid);
			return true, err;
		}
		

		if parseUUIDURL (url, "data/folders", "items", &uuid) {
			err := StorageItemsHandler (db, session, w, r, uuid);
			return true, err;
		}

		if parseUUIDURL (url, "data/items", "entities", &uuid) {
			err := StorageEntitiesHandler (db, session, w, r, uuid);
			return true, err;
		}

		if parseUUIDURL (url, "data/download", "", &uuid) {
			err := StorageDownloadHandler (db, session, w, r, uuid);
			return true, err;
		}
		
	}

	if (r.Method == "POST") {		
		if parseUUIDURL (url, "data/hubs", "", &uuid) {
			err := StorageProjectNewHandler (db, session, w, r, uuid);
			return true, err;
		}
		
		if parseUUIDURL (url, "data/folders", "newfolder", &uuid) {
			err := StorageFolderNewHandler (db, session, w, r, uuid);
			return true, err;
		}

		if parseUUIDURL (url, "data/folders", "newitem", &uuid) {
			err := StorageItemNewHandler (db, session, w, r, uuid);
			return true, err;
		}
		
		if parseUUIDURL (url, "data/upload", "", &uuid) {
			err := StorageBinaryUploadHandler (db, session, w, r, uuid);
			return true, err;
		}

		if parseUUIDURL (url, "data/entities", "", &uuid) {
			err := StorageEntityUpdateHandler (db, session, w, r, uuid);
			return true, err;
		}
					
	}
	
	return false, nil;
	
}
