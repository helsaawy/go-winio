//go:build windows

/*
mkwinerr generates Win32 error definitions for both Windows and other platforms.

It parses files specified on the command line containing error prototypes
and generates separate Windows and non-Windows files.
The Windows-specific files it generates rely on [windows.Errono].


Prototypes are go directives of the form: //err errName = errNumber.
For example:

	//err errInvalidName = 123
*/
package main

func main() {

}
