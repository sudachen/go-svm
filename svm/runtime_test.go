package svm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_RuntimeCreation(t *testing.T) {
	rt, err := NewRuntime()
	assert.NoError(t, err)
	rt.Destroy()
}

func Test_RuntimeValidateCall(t *testing.T) {
	rt, err := NewRuntime()
	assert.NoError(t, err)
	defer rt.Destroy()
	msg, err := CallMessage{0, StringAddress("@target"), "test", []byte{}, []byte{} }.Encode()
	assert.NoError(t, err)
	defer msg.Destroy()
	err = rt.ValidateCall(msg)
	assert.NoError(t, err)
}

func Test_RuntimeDeploy(t *testing.T) {
	rt, err := NewRuntime()
	assert.NoError(t, err)
	defer rt.Destroy()
	msg, err := CallMessage{0, StringAddress("@target"), "test", []byte{}, []byte{} }.Encode()
	assert.NoError(t, err)
	defer msg.Destroy()
	envelope := NewEnvelope(StringAddress("@target"), 100,1,1)
	defer envelope.Destroy()
	ctx := NewContext()
	defer ctx.Destroy()
	_, err = rt.Call(envelope, msg, ctx)
	assert.NoError(t, err)
}

func Test_RuntimeCall(t *testing.T) {
	rt, err := NewRuntime()
	assert.NoError(t, err)
	defer rt.Destroy()
	msg, err := CallMessage{0, StringAddress("@target"), "test", []byte{}, []byte{} }.Encode()
	assert.NoError(t, err)
	defer msg.Destroy()
	envelope := NewEnvelope(StringAddress("@target"), 100,1,1)
	defer envelope.Destroy()
	ctx := NewContext()
	defer ctx.Destroy()
	_, err = rt.Call(envelope, msg, ctx)
	assert.NoError(t, err)
}


func Test_RuntimeSpawn(t *testing.T) {
	rt, err := NewRuntime()
	assert.NoError(t, err)
	defer rt.Destroy()
	msg, err := SpawnMessage{0, StringAddress("@target"), "test", "Ctor", []byte{} }.Encode()
	assert.NoError(t, err)
	defer msg.Destroy()
	envelope := NewEnvelope(StringAddress("@target"), 100,1,1)
	defer envelope.Destroy()
	ctx := NewContext()
	defer ctx.Destroy()
	_, err = rt.Spawn(envelope, msg, ctx)
	assert.NoError(t, err)
}
