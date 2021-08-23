package svm

/*
#include "svm.h"
#include "memory.h"
*/
import "C"

type Context struct {
	ByteArray
}

func NewContext() *Context {
	c := &Context{}
	c.byteArray = C.svm_context_alloc()
	return c
}

