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
	"fmt"
	"log"
	"os"
	"strings"
	"path/filepath"
	"golang.org/x/sys/windows/svc"
)


func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command> [config.xml]\n"+
			"       where <command> is one of\n"+
			"       install, remove, debug, start, stop, pause or continue.\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////
// main function
//////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {	

	
			
	const svcName = "Netfabb Application Service"

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		runService(svcName, false)
		return
	}

	if len(os.Args) < 2 {
		usage("no command specified")
	}
	
	ConfigXMLName := "netfabbapplicationservice.xml";
	if len (os.Args) > 2 {
		ConfigXMLName = os.Args[2];
	}

	Directory, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
    }	
	
	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		runService(svcName, true)
		return
	case "install":
			
		ConfigXMLPath := filepath.Join (Directory, ConfigXMLName);
	
		SaveConfigPathToRegistry (ConfigXMLPath);
	
		err = installService(svcName, "Netfabb Application Service")
	case "remove":
		err = removeService(svcName)
	case "start":
		err = startService(svcName)
	case "stop":
		err = controlService(svcName, svc.Stop, svc.Stopped)
	case "pause":
		err = controlService(svcName, svc.Pause, svc.Paused)
	case "continue":
		err = controlService(svcName, svc.Continue, svc.Running)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, svcName, err)
	}
	return
	
	
}


