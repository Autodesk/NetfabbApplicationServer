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
// netfabbstorage_db.go
// Handles all SQL requests and transfers the data into in memory representations
//////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"database/sql"
	"errors"
	"time"
)


//////////////////////////////////////////////////////////////////////////////////////////////////////
// BeginTransaction
// starts a database transaction
//////////////////////////////////////////////////////////////////////////////////////////////////////

func BeginTransaction (db *sql.DB) (error) {
	_, err := db.Exec ("BEGIN TRANSACTION");
	return err;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// CommittTransaction
// committs a database transaction
//////////////////////////////////////////////////////////////////////////////////////////////////////

func CommittTransaction (db *sql.DB) (error) {
	_, err := db.Exec ("COMMIT");
	return err;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// RollbackTransaction
// rolls back a database transaction
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RollbackTransaction (db *sql.DB) (error) {
	_, err := db.Exec ("ROLLBACK");
	return err;
}



//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveHubs
// retrieves all hubs for a user
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RetrieveHubs (db *sql.DB) ([]NetStorageHub, error) {
	entries := make([] NetStorageHub, 0);
	
	statement, err := db.Prepare ("SELECT uuid, hubname, active FROM netstorage_hubs WHERE active=1");
	if (err != nil) {
		return entries, err;
	}
		
	rows, err := statement.Query();
	if (err != nil) {
		return entries, err;
	}

	for (rows.Next()) {
		var entry NetStorageHub;
		
		err = rows.Scan (&entry.UUID, &entry.Name, &entry.Active);
		if (err != nil) {
			rows.Close ();		
			return entries, err;
		}
								
		entries = append (entries, entry);
	}
	
	rows.Close ();		
	
	return entries, nil;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveProjects
// retrieves all projects for a hub
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RetrieveProjects (db *sql.DB, hubuuid string) ([]NetStorageProject, error) {
	entries := make([] NetStorageProject, 0);
	
	statement, err := db.Prepare ("SELECT uuid, hubuuid, projectname, active FROM netstorage_projects WHERE hubuuid=? AND active=1");
	if (err != nil) {
		return entries, err;
	}
		
	rows, err := statement.Query(hubuuid);
	if (err != nil) {
		return entries, err;
	}

	for (rows.Next()) {
		var entry NetStorageProject;
		err = rows.Scan (&entry.UUID, &entry.HubUUID, &entry.Name, &entry.Active);
		if (err != nil) {
			rows.Close ();		
			return entries, err;
		}
				
		entries = append (entries, entry);
	}
	
	rows.Close ();		
	
	return entries, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveRootFoldersOfProject
// retrieves all root folders of a project
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RetrieveRootFoldersOfProject (db *sql.DB, projectuuid string) ([]NetStorageFolder, error) {
	entries := make([] NetStorageFolder, 0);
	
	statement, err := db.Prepare ("SELECT uuid, projectuuid, parentuuid, foldername, active FROM netstorage_folders WHERE projectuuid=? AND parentuuid=\"\" AND active=1");
	if (err != nil) {
		return entries, err;
	}
		
	rows, err := statement.Query(projectuuid);
	if (err != nil) {
		return entries, err;
	}

	for (rows.Next()) {
		var entry NetStorageFolder;		
		err = rows.Scan (&entry.UUID, &entry.ProjectUUID, &entry.ParentUUID, &entry.Name, &entry.Active);
		if (err != nil) {
			rows.Close ();		
			return entries, err;
		}
								
		entries = append (entries, entry);
	}
	
	rows.Close ();		
	
	return entries, nil;
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveSubFoldersOfFolder
// retrieves all subfolders of a folder
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RetrieveSubFoldersOfFolder (db *sql.DB, folderuuid string) ([]NetStorageFolder, error) {
	entries := make([] NetStorageFolder, 0);
	
	statement, err := db.Prepare ("SELECT uuid, projectuuid, parentuuid, foldername, active FROM netstorage_folders WHERE parentuuid=? AND active=1");
	if (err != nil) {
		return entries, err;
	}
		
	rows, err := statement.Query(folderuuid);
	if (err != nil) {
		return entries, err;
	}

	for (rows.Next()) {
		var entry NetStorageFolder;
	
		err = rows.Scan (&entry.UUID, &entry.ProjectUUID, &entry.ParentUUID, &entry.Name, &entry.Active);
		if (err != nil) {
			rows.Close ();		
			return entries, err;
		}
				
		entries = append (entries, entry);
	}
	
	rows.Close ();		
	
	return entries, nil;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveFolderByUUID
// retrieves a folder by uuid
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RetrieveFolderByUUID (db *sql.DB, folderuuid string) (NetStorageFolder, error) {
	var folder NetStorageFolder;
	
	statement, err := db.Prepare ("SELECT uuid, projectuuid, parentuuid, foldername, active FROM netstorage_folders WHERE uuid=? AND active=1");
	if (err != nil) {
		return folder, err;
	}
		
	rows, err := statement.Query(folderuuid);
	if (err != nil) {
		return folder, err;
	}

	if (!rows.Next()) {
	    rows.Close();
		return folder, errors.New("folder not found: " + folderuuid);		
	}	
	
	err = rows.Scan (&folder.UUID, &folder.ProjectUUID, &folder.ParentUUID, &folder.Name, &folder.Active);
	rows.Close ();		
	
	return folder, nil;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveItems
// retrieves all items of a folder
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RetrieveItemsOfFolder (db *sql.DB, folderuuid string) ([]NetStorageItem, error) {
	entries := make([] NetStorageItem, 0);
	
	statement, err := db.Prepare ("SELECT netstorage_items.uuid, netstorage_items.folderuuid, netstorage_folders.projectuuid, netstorage_items.itemname, netstorage_items.active FROM netstorage_items LEFT JOIN netstorage_folders ON netstorage_folders.uuid=netstorage_items.folderuuid WHERE folderuuid=? AND netstorage_items.active=1");
	if (err != nil) {
		return entries, err;
	}
		
	rows, err := statement.Query(folderuuid);
	if (err != nil) {
		return entries, err;
	}

	for (rows.Next()) {
		var entry NetStorageItem;
	
		err = rows.Scan (&entry.UUID, &entry.FolderUUID, &entry.ProjectUUID, &entry.Name, &entry.Active);
		if (err != nil) {
			rows.Close ();		
			return entries, err;
		}
								
		entries = append (entries, entry);
	}
	
	rows.Close ();		
	
	return entries, nil;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveItemByUUID
// retrieves a item by uuid
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RetrieveItemByUUID (db *sql.DB, itemuuid string) (NetStorageItem, error) {
	var item NetStorageItem;
	
	statement, err := db.Prepare ("SELECT netstorage_items.uuid, netstorage_items.folderuuid, netstorage_folders.projectuuid, netstorage_items.itemname, netstorage_items.active FROM netstorage_items LEFT JOIN netstorage_folders ON netstorage_folders.uuid=netstorage_items.folderuuid WHERE netstorage_items.uuid=? AND netstorage_items.active=1");
	if (err != nil) {
		return item, err;
	}
		
	rows, err := statement.Query(itemuuid);
	if (err != nil) {
		return item, err;
	}

	if (!rows.Next()) {
	    rows.Close();
		return item, errors.New("item not found: " + itemuuid);		
	}	
	
	err = rows.Scan (&item.UUID, &item.FolderUUID, &item.ProjectUUID, &item.Name, &item.Active);
	rows.Close ();		
	
	return item, nil;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveEntityByUUID
// retrieves a entity by uuid
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RetrieveEntitiesOfItem (db *sql.DB, itemuuid string) ([]NetStorageEntity, error) {
	var entities []NetStorageEntity;
	
	statement, err := db.Prepare ("SELECT netstorage_entities.uuid, netstorage_entities.itemuuid, netstorage_entities.datatype, netstorage_entities.sha1, netstorage_entities.filesize, netstorage_entities.metadata, netstorage_entities.timestamp, netstorage_entities.active FROM netstorage_entities WHERE netstorage_entities.itemuuid=? ORDER BY timestamp");
	if (err != nil) {
		return entities, err;
	}
		
	rows, err := statement.Query(itemuuid);
	if (err != nil) {
		return entities, err;
	}
	
	defer rows.Close();


	for (rows.Next()) {
		var entity NetStorageEntity;
		
		err = rows.Scan (&entity.UUID, &entity.ItemUUID, &entity.DataType, &entity.SHA1, &entity.FileSize, &entity.MetaData, &entity.TimeStamp, &entity.Active);
		if (err != nil) {
			return entities, err;
		}
		
		
		entities = append (entities, entity);
	}
	
	
	return entities, nil;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// RetrieveEntityByUUID
// retrieves a entity by uuid
//////////////////////////////////////////////////////////////////////////////////////////////////////

func RetrieveEntityByUUID (db *sql.DB, entityuuid string, needstobeactive bool) (NetStorageEntity, error) {
	var entity NetStorageEntity;

	activecondition	:= "";	
	if (needstobeactive) {
		activecondition = " AND netstorage_entities.active=1"
	}
	
	statement, err := db.Prepare ("SELECT netstorage_entities.uuid, netstorage_entities.itemuuid, netstorage_entities.datatype, netstorage_entities.sha1, netstorage_entities.filesize, netstorage_entities.metadata, netstorage_entities.active FROM netstorage_entities WHERE netstorage_entities.uuid=?" + activecondition);
	if (err != nil) {
		return entity, err;
	}
		
	rows, err := statement.Query(entityuuid);
	if (err != nil) {
		return entity, err;
	}
	
	defer rows.Close();

	if (!rows.Next()) {
		return entity, errors.New("entity not found: " + entityuuid);		
	}	
	
	err = rows.Scan (&entity.UUID, &entity.ItemUUID, &entity.DataType, &entity.SHA1, &entity.FileSize, entity.MetaData, &entity.Active);
	
	return entity, nil;
}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// createNewProject
// creates a new project DB entry
//////////////////////////////////////////////////////////////////////////////////////////////////////

func createNewProject (db *sql.DB, projectuuid string, projectname string, hubuuid string) (error) {

	statement1, err := db.Prepare ("SELECT uuid, hubname FROM netstorage_hubs WHERE uuid=? AND active=1");
	if (err != nil) {
		return err;
	}
		
	rows, err := statement1.Query(hubuuid);
	if (err != nil) {
		return err;
	}

	hasRow := rows.Next();
	rows.Close ();		
		
	if (!hasRow) {
		return errors.New("hub not found!");
	}


	statement2, err := db.Prepare ("INSERT INTO netstorage_projects (uuid, projectname, hubuuid, active) VALUES (?, ?, ?, 1)");
	if (err != nil) {
		return err;
	}
	
	_, err = statement2.Exec(projectuuid, projectname, hubuuid);	
	return err;

}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// createNewFolder
// creates a new folder DB entry
//////////////////////////////////////////////////////////////////////////////////////////////////////

func createNewFolder (db *sql.DB, folderuuid string, projectuuid string, foldername string, parentuuid string) (error) {

	statement1, err := db.Prepare ("SELECT uuid FROM netstorage_projects WHERE uuid=? AND active=1");
	if (err != nil) {
		return err;
	}
		
	rows, err := statement1.Query(projectuuid);
	if (err != nil) {
		return err;
	}

	hasRow := rows.Next();
	rows.Close ();		
		
	if (!hasRow) {
		return errors.New("project not found!");
	}

	if (parentuuid != "") {
		statement2, err := db.Prepare ("SELECT uuid FROM netstorage_folders WHERE uuid=? AND active=1");
		if (err != nil) {
			return err;
		}
			
		rows, err := statement2.Query(parentuuid);
		if (err != nil) {
			return err;
		}

		hasRow := rows.Next();
		rows.Close ();		
			
		if (!hasRow) {
			return errors.New("parent folder not found!");
		}
	}

	
	statement3, err := db.Prepare ("INSERT INTO netstorage_folders (uuid, foldername, projectuuid, parentuuid, active) VALUES (?, ?, ?, ?, 1)");
	if (err != nil) {
		return err;
	}
	
	_, err = statement3.Exec(folderuuid, foldername, projectuuid, parentuuid);	
	return err;

}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// createNewItem
// creates a new item DB entry
//////////////////////////////////////////////////////////////////////////////////////////////////////

func createNewItem (db *sql.DB, itemuuid string, folderuuid string, itemname string) (error) {
	
	
	statement1, err := db.Prepare ("INSERT INTO netstorage_items (uuid, itemname, folderuuid, active) VALUES (?, ?, ?, 1)");
	if (err != nil) {
		return err;
	}
	
	_, err = statement1.Exec(itemuuid, itemname, folderuuid);	
	return err;

}


//////////////////////////////////////////////////////////////////////////////////////////////////////
// createNewEntity
// creates a new entity DB entry
//////////////////////////////////////////////////////////////////////////////////////////////////////

func createNewEntity (db *sql.DB, entityuuid string, itemuuid string, sha1 string, filesize int64, active bool) (error) {

	var activedbfield int;
	if (active) {
		activedbfield = 1;
	} else {
		activedbfield = 0;
	} 
	
	timestamp := time.Now().Format(time.RFC3339);
				
	statement1, err := db.Prepare ("INSERT INTO netstorage_entities (uuid, itemuuid, sha1, filesize, timestamp, active) VALUES (?, ?, ?, ?, ?, ?)");
	if (err != nil) {
		return err;
	}
		
	_, err = statement1.Exec(entityuuid, itemuuid, sha1, filesize, timestamp, activedbfield);	
	
	return err;

}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// updateEntityStatus
// updates activity status of an db entity
//////////////////////////////////////////////////////////////////////////////////////////////////////

func updateEntity (db *sql.DB, entityuuid string, datatype string, metadata string, active bool) (error) {

	var activedbfield int;
	if (active) {
		activedbfield = 1;
	} else {
		activedbfield = 0;
	}
				
	statement1, err := db.Prepare ("UPDATE netstorage_entities SET active=?, datatype=?, metadata=? WHERE uuid=?");
	if (err != nil) {
		return err;
	}
	
	_, err = statement1.Exec(activedbfield, datatype, metadata, entityuuid);	
	return err;

}

