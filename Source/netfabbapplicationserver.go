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
	"os"
	"fmt"
	"log"
	
	
)


//////////////////////////////////////////////////////////////////////////////////////////////////////
// main function
//////////////////////////////////////////////////////////////////////////////////////////////////////

func main() {	

	fmt.Println ("");                                        
	fmt.Println ("  .';::;'.              .oxddol:. ");
	fmt.Println (" .:oodxkkd:.            'xOkkkxk:   ___        _            _           _     ");
	fmt.Println (" ,clodxkkOOx:.          'xOkkkkk;  / _ \\      | |          | |         | |   ");
	fmt.Println (" ;llodxkkOOO0x,         'xOkkkkk; / /_\\ \\_   _| |_ ___   __| | ___  ___| | _");
	fmt.Println (" ,cldddddxkO000d'       .xOkkxxk; |  _  | | | | __/ _ \\ / _` |/ _ \\/ __| |/ /");
	fmt.Println (".:codoc:::coxO00Ol;.    .xOkxxxx; | | | | |_| | || (_) | (_| |  __/\\__ \\   < ");
	fmt.Println (".:loolllccc:cdO0000x:.  .dkxxxxx, \\_| |_/\\__,_|\\__\\___/ \\__,_|\\___||___/_|\\_\\");
	fmt.Println (".:odooolllllccok0OOOOd, 'dkxxxxo. ");
	fmt.Println (".:oddooooclollcokOOOOOkooxxxxdxo. ");
	fmt.Println (".cdddddxl..,lolloxOkOOkkkxddxdxo.  _   _      _    __      _     _          ");
	fmt.Println (".cdxdddx:   .:oollooodxxxxddddxo. | \\ | |    | |  / _|    | |   | |       ");
	fmt.Println (".lxxxxxd,     .:lllllccloooooddo' |  \\| | ___| |_| |_ __ _| |__ | |__     ");
	fmt.Println (".oxxxxxx,       .;clllcccllooodd' | . ` |/ _ \\ __|  _/ _` | '_ \\| '_ \\  ");
	fmt.Println (".lxxxxxx,          .;clccclloodl. | |\\  |  __/ |_| || (_| | |_) | |_) |  ");
	fmt.Println (" :kkkkkx,            .;llcllooo,  \\_| \\_/\\___|\\__|_| \\__,_|_.__/|_.__/");
	fmt.Println (" cOkkkkk;              .;llc:,.   ");
	fmt.Println (" .;;;;;,.                ...      ");
	fmt.Println ("");                                      
	fmt.Println ("");	

	log.Println(fmt.Sprintf("%s (%s)", CONFIG_NAME, CONFIG_VERSION));		

	
	// Handle REST endpoint	
	ConfigFileName := "netfabbapplicationserver.xml";
	if (len (os.Args) > 1) {
		ConfigFileName = os.Args[1];
	}
			
	err := startAppServer (ConfigFileName, true);
	if (err != nil) {
		log.Fatal (err);
	}		
	
	
}


