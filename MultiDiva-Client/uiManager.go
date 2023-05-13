package main

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"

func setUIString(stringToSet *C.char, newString string, CStringLength C.ulonglong) {
	C.strncpy((*C.char)(stringToSet), (*C.char)(C.CString(newString)), CStringLength)
}
