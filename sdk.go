package hypertune

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

/*
#cgo arm64 LDFLAGS: -L${SRCDIR}/lib/aarch64 -lhypertune
#cgo amd64 LDFLAGS: -L${SRCDIR}/lib/x86_64 -lhypertune
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include "lib/hypertune.h"
*/
import "C"

const language = "go"

type CreateConfig struct {
	Token                *string
	VariableValues       any
	QueryJSON            string
	InitQueryJSON        string
	FallbackInitDataJSON *string
	Options              []Option
}

func Create(args CreateConfig) (*Node, error) {
	configJSON, err := json.Marshal(parseOptions(args.Options))
	if err != nil {
		return nil, fmt.Errorf(`failed to marshal sdk config: %w`, err)
	}

	varValuesJSON, err := json.Marshal(args.VariableValues)
	if err != nil {
		return nil, fmt.Errorf(`failed to marshal sdk config: %w`, err)
	}

	varValuesCStr := C.CString(string(varValuesJSON))
	defer C.free(unsafe.Pointer(varValuesCStr))

	var fallbackCStr *C.char
	if args.FallbackInitDataJSON != nil {
		fallbackCStr = C.CString(*args.FallbackInitDataJSON)
		defer C.free(unsafe.Pointer(fallbackCStr))
	}

	var tokenCStr *C.char
	if args.Token != nil {
		tokenCStr = C.CString(*args.Token)
		defer C.free(unsafe.Pointer(tokenCStr))
	}

	initQueryCStr := C.CString(args.InitQueryJSON)
	defer C.free(unsafe.Pointer(initQueryCStr))

	queryCStr := C.CString(args.QueryJSON)
	defer C.free(unsafe.Pointer(queryCStr))

	configCStr := C.CString(string(configJSON))
	defer C.free(unsafe.Pointer(configCStr))

	return newNode(C.create(
		varValuesCStr,
		fallbackCStr,
		tokenCStr,
		initQueryCStr,
		queryCStr,
		configCStr,
	)), nil
}
