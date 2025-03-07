package hypertune

import (
	"encoding/json"
	"fmt"
	"iter"
	"runtime"
	"unsafe"
)

/*
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include "lib/hypertune.h"
*/
import "C"

type NodeType uint32

const (
	NodeTypeString NodeType = iota
	NodeTypeEnum
	NodeTypeInt
	NodeTypeFloat
	NodeTypeBool
	NodeTypeObject
	NodeTypeVoid
	NodeTypeUnknown
)

func IsStringNodeType(nodeType NodeType) bool {
	return nodeType == NodeTypeString || nodeType == NodeTypeEnum
}

func IsNumberNodeType(nodeType NodeType) bool {
	return nodeType == NodeTypeFloat || nodeType == NodeTypeInt
}

type NodeIteratorState uint32

const (
	NodeIteratorStateMaybeMore NodeIteratorState = iota
	NodeIteratorStateConsumed
)

type Node struct {
	handle         C.uint32_t
	error          bool
	Type           NodeType
	EnumValue      *string
	ObjectTypeName *string
}

func newNode(result C.NodeResult) *Node {
	props := &Node{
		handle: result.id,
		error:  bool(result.error),
		Type:   NodeType(result._type),
	}

	if !props.error {
		props.EnumValue = readSizedString(result.enum_value)
		props.ObjectTypeName = readSizedString(result.object_type_name)
	}

	runtime.SetFinalizer(props, func(n *Node) {
		C.node_free(n.handle)
	})

	return props
}

func newNodeIterator(iterator C.NodeIteratorResult) iter.Seq[*Node] {
	return func(yield func(*Node) bool) {
		if iterator.error {
			return
		}
		for {
			next_result := C.node_iterator_next(C.uint32_t(iterator.id))
			if NodeIteratorState(next_result.state) == NodeIteratorStateConsumed {
				return
			}
			if next_result.error {
				return
			}
			if !yield(newNode(next_result.node)) {
				return
			}
		}
	}
}

func (n *Node) Close() {
	C.node_close(n.handle)
}

func (n *Node) WaitForInitialization() {
	C.wait_for_initialization(n.handle)
}

func (n *Node) FlushLogs() {
	C.node_flush_logs(n.handle)
}

func (n *Node) GetField(field string, arguments any) *Node {
	argumentsJSON, err := json.Marshal(arguments)
	if err != nil {
		panic(fmt.Errorf("failed to field marshal arguments for field %s: %w", field, err))
	}
	return n.GetFieldWithJSONArguments(field, string(argumentsJSON))
}

func (n *Node) GetFieldWithJSONArguments(field string, argumentsJSON string) *Node {
	fieldCStr := C.CString(field)
	defer C.free(unsafe.Pointer(fieldCStr))

	argumentsJSONCStr := C.CString(argumentsJSON)
	defer C.free(unsafe.Pointer(argumentsJSONCStr))

	return newNode(C.node_get_field(
		n.handle,
		fieldCStr,
		argumentsJSONCStr,
	))
}

func (n *Node) GetItems() iter.Seq[*Node] {
	return newNodeIterator(C.node_get_items(n.handle))
}

func (n *Node) Evaluate() (any, error) {
	result := C.node_evaluate(n.handle)
	if result.error {
		return nil, fmt.Errorf("evaluation error")
	}

	if result.value == nil || result.length == 0 {
		return nil, nil
	}

	value := C.GoStringN((*C.char)(unsafe.Pointer(result.value)), C.int(result.length))
	var parsed any
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

func (n *Node) LogUnexpectedTypeError() {
	C.node_log_unexpected_type_error(n.handle)
}

func (n *Node) LogUnexpectedValueError(value any) {
	C.node_log_unexpected_value_error(n.handle)
}

func readSizedString(str C.SizedString) *string {
	if str.length == 0 {
		return nil
	}
	result := string(str.bytes[:str.length])

	return &result
}
