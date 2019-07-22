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


package main

import (
	"encoding/json"
	"sync"
)

// Protocol Constants

const PROTOCOL_VERSION = "2.0.0"
const PROTOCOL_ERROR = "com.autodesk.error"

const PROTOCOL_NEWSESSION = "com.autodesk.netfabbsession.new"
const PROTOCOL_CLOSESESSION = "com.autodesk.netfabbsession.close"
const PROTOCOL_AUTHSESSION = "com.autodesk.netfabbsession.auth"

const PROTOCOL_HUBS = "com.autodesk.netfabbstorage.hubs"
const PROTOCOL_PROJECTS = "com.autodesk.netfabbstorage.projects"
const PROTOCOL_ROOTFOLDERS = "com.autodesk.netfabbstorage.rootfolders"
const PROTOCOL_SUBFOLDERS = "com.autodesk.netfabbstorage.subfolders"
const PROTOCOL_ITEMS = "com.autodesk.netfabbstorage.items"
const PROTOCOL_ENTITIES = "com.autodesk.netfabbstorage.entities"
const PROTOCOL_NEWPROJECT = "com.autodesk.netfabbstorage.newproject"
const PROTOCOL_NEWFOLDER = "com.autodesk.netfabbstorage.newfolder"
const PROTOCOL_NEWITEM = "com.autodesk.netfabbstorage.newitem"
const PROTOCOL_NEWENTITY = "com.autodesk.netfabbstorage.newentity"
const PROTOCOL_UPDATEENTITY = "com.autodesk.netfabbstorage.updateentity"

const PROTOCOL_ORMREAD = "com.autodesk.netfabborm.read"
const PROTOCOL_ORMSAVE = "com.autodesk.netfabborm.save"
const PROTOCOL_ORMCREATE = "com.autodesk.netfabborm.create"
const PROTOCOL_ORMDELETE = "com.autodesk.netfabborm.delete"
const PROTOCOL_ORMUPDATE = "com.autodesk.netfabborm.update"
const PROTOCOL_ORMSCHEMA = "com.autodesk.netfabborm.schema"

const PROTOCOL_TASKNEW = "com.autodesk.netfabbtasks.new"
const PROTOCOL_TASKCLEAR = "com.autodesk.netfabbtasks.clear"
const PROTOCOL_TASKUPDATE = "com.autodesk.netfabbtasks.update"
const PROTOCOL_TASKHANDLE = "com.autodesk.netfabbtasks.handle"
const PROTOCOL_TASKSTATUS = "com.autodesk.netfabbtasks.status"


// Data Structures

type NetStorageSession struct {
    UUID string `json:"uuid"`
    LogUUID string `json:"loguuid"`
	LogIndex int `json:"logindex"`
    Token string `json:"token"`
    UserID string `json:"userid"`
    Active int `json:"active"`
	Mutex sync.Mutex
}


type NetStorageToken struct {
    SessionUUID string `json:"session"`
    UserID string `json:"userid"`
}


type NetStorageHub struct {
    UUID string `json:"uuid"`
    Name string `json:"name"`
    Active int `json:"active"`
}

type NetStorageProject struct {
    UUID string `json:"uuid"`
    HubUUID string `json:"hubuuid"`
    Name string `json:"name"`
    Active int `json:"active"`
}

type NetStorageFolder struct {
    UUID string `json:"uuid"`
    ProjectUUID string `json:"projectuuid"`
    ParentUUID string `json:"parentuuid"`
    Name string `json:"name"`
    Active int `json:"active"`
}

type NetStorageItem struct {
    UUID string `json:"uuid"`
    ProjectUUID string `json:"projectuuid"`
    FolderUUID string `json:"folderuuid"`
    Name string `json:"name"`
    Active int `json:"active"`
}

type NetStorageEntity struct {
    UUID string `json:"uuid"`
    ItemUUID string `json:"itemuuid"`
	DataType string `json:"datatype"`
	SHA1 string `json:"sha1"`
	FileSize string `json:"filesize"`
	MetaData string `json:"metadata"`
	TimeStamp string `json:"timestamp"`
    Active int `json:"active"`
}


// Protocol Header

type NetStorageProtocolHeader struct {
	Protocol string `json:"protocol"`
	Version string `json:"version"`
}

type NetStorageHeaderInterface interface {
	GetHeader() NetStorageProtocolHeader
}


// Request protocol

type NetStorageCreateSessionRequest struct {
	NetStorageProtocolHeader
    UserID string `json:"userid"`
}

type NetStorageCloseSessionRequest struct {
	NetStorageProtocolHeader
    SessionUUID string `json:"sessionuuid"`
}

type NetStorageAuthenticateSessionRequest struct {
	NetStorageProtocolHeader
    SessionUUID string `json:"sessionuuid"`
    AuthType string `json:"authtype"`
    AuthKey string `json:"authkey"`
}


type NetStorageNewProjectRequest struct {
	NetStorageProtocolHeader
    ProjectName string `json:"projectname"`
}

type NetStorageNewFolderRequest struct {
	NetStorageProtocolHeader
    FolderName string `json:"foldername"`
}

type NetStorageNewItemRequest struct {
	NetStorageProtocolHeader
    ItemName string `json:"itemname"`
}

type NetStorageUpdateEntityRequest struct {
	NetStorageProtocolHeader
    DataType string `json:"datatype"`
	MetaData json.RawMessage `json:"metadata"`
}


// ORM schemas

type NetORMValue struct {
    Key string `json:"key"`
    Type string `json:"type"`
    Value string `json:"value"`
	Unique bool `json:"unique"`
}

type NetORMRow []string;

type NetORMTableMapping struct {
    Name string
    PrimaryKey string 
	FieldMap map[string]NetORMValue
}


type NetORMMapping struct {
    Name string `json:"name"`
    PrimaryKey string `json:"primarykey"`
	Fields []NetORMValue `json:"fields"`
}


type NetORMSchema struct {
    Schema string `json:"schema"`
    Version string `json:"version"`
	Mappings []NetORMMapping `json:"mappings"`
}


// ORM protocol

type NetORMReadRequest struct {
	NetStorageProtocolHeader
    Entity string `json:"entity"`
	Values []NetORMValue `json:"values"`
	Filter []NetORMValue `json:"filter"`
}

type NetORMSaveRequest struct {
	NetStorageProtocolHeader
    Entity string `json:"entity"`
	Values []NetORMValue `json:"values"`
}

type NetORMDeleteRequest struct {
	NetStorageProtocolHeader
    Entity string `json:"entity"`
	Filter []NetORMValue `json:"filter"`
}

type NetORMUpdateRequest struct {
	NetStorageProtocolHeader
    Entity string `json:"entity"`
	Values []NetORMValue `json:"values"`
	Filter []NetORMValue `json:"filter"`
}

// Task Protocol
type NetTaskNewRequest struct {
	NetStorageProtocolHeader
	Name string `json:"name"`
	Parameters map[string]string `json:"parameters"`
}

type NetTaskClearRequest struct {
	NetStorageProtocolHeader
}

type NetTaskHandleRequest struct {
	NetStorageProtocolHeader
	Name string `json:"name"`
	Worker string `json:"worker"`
}

type NetTaskUpdateRequest struct {
	NetStorageProtocolHeader
	Status string `json:"status"`
	WorkerSecret string `json:"workersecret"`
	Results map[string]string `json:"results"`
}

type NetTaskStatusRequest struct {
	NetStorageProtocolHeader
	UUID string `json:"uuid"`
}


// Request interfaces

func (request *NetStorageCreateSessionRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetStorageCloseSessionRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetStorageAuthenticateSessionRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}


func (request *NetStorageNewProjectRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetStorageNewFolderRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetStorageNewItemRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetStorageUpdateEntityRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetORMReadRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetORMSaveRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetORMDeleteRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetORMUpdateRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetTaskNewRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetTaskClearRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetTaskHandleRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetTaskUpdateRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetTaskStatusRequest) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

// Reply protocol

type NetStorageErrorReply struct {
	NetStorageProtocolHeader
	ErrorMessage string `json:"errormessage"`
	LogUUID string `json:"loguuid"`
}


type NetStorageCreateSessionReply struct {
	NetStorageProtocolHeader
    SessionUUID string `json:"sessionuuid"`
    AuthType string `json:"authtype"`
    UserID string `json:"userid"`
    Salt string `json:"salt"`
}

type NetStorageCloseSessionReply struct {
	NetStorageProtocolHeader
    SessionUUID string `json:"sessionuuid"`
}

type NetStorageAuthenticateSessionReply struct {
	NetStorageProtocolHeader
    SessionUUID string `json:"sessionuuid"`
    Token string `json:"token"`
}


type NetStorageHubReply struct {
	NetStorageProtocolHeader
	Hubs []NetStorageHub `json:"hubs"`
}

type NetStorageProjectsReply struct {
	NetStorageProtocolHeader
    HubUUID string `json:"hubuuid"`
	Projects []NetStorageProject `json:"projects"`
}

type NetStorageFoldersReply struct {
	NetStorageProtocolHeader
	Folders []NetStorageFolder `json:"folders"`
}

type NetStorageItemsReply struct {
	NetStorageProtocolHeader
	Items []NetStorageItem `json:"items"`
}

type NetStorageEntitiesReply struct {
	NetStorageProtocolHeader
	Entities []NetStorageEntity `json:"entities"`
}

type NetStorageNewProjectReply struct {
	NetStorageProtocolHeader
    HubUUID string `json:"hubuuid"`
    ProjectUUID string `json:"projectuuid"`
    RootFolderUUID string `json:"rootfolderuuid"`
}

type NetStorageNewFolderReply struct {
	NetStorageProtocolHeader
    ProjectUUID string `json:"projectuuid"`
    ParentUUID string `json:"parentuuid"`
    SubFolderUUID string `json:"subfolderuuid"`
}

type NetStorageNewItemReply struct {
	NetStorageProtocolHeader
    ItemUUID string `json:"itemuuid"`
    FolderUUID string `json:"folderuuid"`
}

type NetStorageNewEntityReply struct {
	NetStorageProtocolHeader
    ItemUUID string `json:"itemuuid"`
    EntityUUID string `json:"entityuuid"`
}

type NetStorageUpdateEntityReply struct {
	NetStorageProtocolHeader
    ItemUUID string `json:"itemuuid"`
    EntityUUID string `json:"entityuuid"`
}


// ORM protocol

type NetORMReadReply struct {
	NetStorageProtocolHeader
	Columns []string `json:"columns"`
	Rows []NetORMRow `json:"rows"`
}

type NetORMSaveReply struct {
	NetStorageProtocolHeader
}

type NetORMDeleteReply struct {
	NetStorageProtocolHeader
}

type NetORMUpdateReply struct {
	NetStorageProtocolHeader
}


// Task protocol
type NetTaskNewReply struct {
	NetStorageProtocolHeader
	UUID string `json:"uuid"`
}

type NetTaskClearReply struct {
	NetStorageProtocolHeader
	Count int `json:"count"`
}

type NetTaskHandleReply struct {
	NetStorageProtocolHeader
	UUID string `json:"uuid"`
	WorkerSecret string `json:"workersecret"`
	Name string `json:"name"`
	Parameters map[string]string `json:"parameters"`
}

type NetTaskUpdateReply struct {
	NetStorageProtocolHeader
	UUID string `json:"uuid"`
}

type NetTaskStatusReply struct {
	NetStorageProtocolHeader
	UUID string `json:"uuid"`
	Status string `json:"status"`
	Name string `json:"name"`
	Parameters map[string]string `json:"parameters"`
    Result map[string]string `json:"result"`
	Worker string `json:"worker"`
	TimeStamp string `json:"timestamp"`
}




func (request *NetTaskNewReply) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetTaskClearReply) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetTaskHandleReply) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetTaskUpdateReply) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}

func (request *NetTaskStatusReply) GetHeader() NetStorageProtocolHeader {
  return request.NetStorageProtocolHeader;
}


