# Netfabb Application Server
The "Netfabb Application Server" (short: NAS) is a server (or: service) program written in Go, exposing a RESTful webservice API to client programs for a central storage system and distributed task management.

## Current features:
* A Storage Server part that connects netfabb Desktop clients together (the Netfabb Desktop software contains already a client for it)
* A Task manager feature allowing to distribute jobs among multiple clients
* A basic authentication concept with usernames/salted password and against the Autodesk Authentication service is implemented
* Logging and persistence are provided with a SQLLite database
* The central services are working in dialog with an Autodesk Netfabb client implementation

## How to use NAS:
As a starting point please look into the "doc" directory.

## Contributing
NAS is an open source project.
Contributions are welcome and we are looking for people that can improve existing functionality or create new integrations. Have a look at the [contributor's guide](CONTRIBUTING.md) for details.

## License and 3rd party acknowledgements
* NAS has a [BSD-3-Clause license](LICENSE.md)
* NAS uses these [3rd Party components](3RD_PARTY.md)