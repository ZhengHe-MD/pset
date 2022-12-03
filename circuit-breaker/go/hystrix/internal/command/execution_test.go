package command

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewExecution(t *testing.T) {
	e := NewExecution()
	assert.Equal(t, ExecutionStatusUnspecified, e.Status)
	assert.Equal(t, ExecutionStatusUnspecified, e.FallbackStatus)
	assert.Equal(t, time.Time{}, e.start)
	assert.Equal(t, time.Duration(0), e.Duration)
}

func TestExecution_Start_Finish(t *testing.T) {
	e := NewExecution()
	e.Start()
	time.Sleep(time.Second)
	e.Finish()
	assert.Equal(t, int64(time.Second.Seconds()), int64(e.Duration.Seconds()))
}
