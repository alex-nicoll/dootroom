package main

import (
	"strings"
	"testing"

	"dootroom.com/main/internal"
)

func Test_startPipelineInternal(t *testing.T) {
	unmarshalOut := make(chan interface{})
	modelChan := make(chan interface{})
	handleConn := startPipelineInternal(unmarshalOut, modelChan)

	// When a new connection is made, the pipeline should send the Game of Life
	// state as JSON to that connection.

	conn1 := &internal.MockConn{
		In:  make(chan []byte),
		Out: make(chan []byte),
	}
	conn2 := &internal.MockConn{
		In:  make(chan []byte),
		Out: make(chan []byte),
	}

	handleConn(conn1)
	handleConn(conn2)

	json := string(<-conn1.Out)
	if !strings.HasPrefix(json, "[") {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(<-conn2.Out)
	if !strings.HasPrefix(json, "[") {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// When two Game of Life diffs (that is, the difference between the current
	// state and the desired state) come in as JSON on separate connections and
	// then a tick occurs, the pipeline should send the combined diff to each
	// connection.

	diff1 := "{\"30\":{\"30\":true},\"31\":{\"32\":true}}"
	diff2 := "{\"30\":{\"30\":true,\"31\":true},\"31\":{\"31\":true},\"32\":{\"31\":true}}"
	combinedDiff := "{\"30\":{\"30\":true,\"31\":true},\"31\":{\"31\":true,\"32\":true},\"32\":{\"31\":true}}"

	conn1.In <- []byte(diff1)
	conn2.In <- []byte(diff2)
	modelChan <- <-unmarshalOut
	modelChan <- <-unmarshalOut
	modelChan <- &Tick{}

	json = string(<-conn1.Out)
	if json != combinedDiff {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(<-conn2.Out)
	if json != combinedDiff {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// When a tick occurs, the pipeline should send the diff between the
	// current state and the previous state to each connection.

	diff := "{\"30\":{\"32\":true},\"31\":{\"31\":false},\"32\":{\"32\":true}}"

	modelChan <- &Tick{}

	json = string(<-conn1.Out)
	if json != diff {
		t.Errorf("Got incorrect JSON: %v", json)
	}
	json = string(<-conn2.Out)
	if json != diff {
		t.Errorf("Got incorrect JSON: %v", json)
	}

	// When a new connection is made, the pipeline should send the Game of Life
	// state as JSON to that connection. Unlike in the previous case, there
	// should now be non-empty cells.

	conn3 := &internal.MockConn{
		In:  make(chan []byte),
		Out: make(chan []byte),
	}

	handleConn(conn3)

	json = string(<-conn3.Out)
	if !strings.HasPrefix(json, "[") || !strings.Contains(json, "true") {
		t.Errorf("Got incorrect JSON: %v", json)
	}
}
