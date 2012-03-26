package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Abstract bulk update instruction.
type Instruction interface {
	writeTo(w io.Writer) error
}

// Instruction to update an index entry.
type UpdateInstruction struct {
	Id      string                 `json:"_id"`
	Index   string                 `json:"_index"`
	Type    string                 `json:"_type"`
	Routing string                 `json:"_routing,omitempty"`
	Body    map[string]interface{} `json:"-"`
}

func (ui *UpdateInstruction) writeTo(w io.Writer) error {
	e := json.NewEncoder(w)
	err := e.Encode(map[string]interface{}{
		"index": ui,
	})
	if err != nil {
		return err
	}
	err = e.Encode(ui.Body)
	return err
}

// Instruction to delete an item from an index.
type DeleteInstruction struct {
	Id      string `json:"_id"`
	Index   string `json:"_index"`
	Type    string `json:"_type"`
	Routing string `json:"_routing,omitempty"`
}

func (di *DeleteInstruction) writeTo(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(map[string]interface{}{
		"delete": di,
	})
}

type bulkWriter struct {
	update chan Instruction
	reqch  chan chan *http.Request
	quit   chan bool
	w      *bytes.Buffer
}

// Interface for writing bulk data into elasticsearch.
type BulkUpdater interface {
	// Update the index with a new record (or delete a record).
	Update(ui Instruction)
	// Send the current batch.
	SendBatch() error
	// Shut down this bulk interface
	Quit()
}

func (b *bulkWriter) Update(ui Instruction) {
	b.update <- ui
}

func (b *bulkWriter) SendBatch() error {
	reqch := make(chan *http.Request)
	b.reqch <- reqch
	req := <-reqch

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// TODO: Parse the response and check each thingy.
	if resp.StatusCode > 201 {
		return errors.New("HTTP error:  " + resp.Status)
	}
	return nil
}

func (b *bulkWriter) Quit() {
	b.quit <- true
}

func issueBulkRequest(u string, bw *bulkWriter, reqch chan *http.Request) {
	req, err := http.NewRequest("POST", u+"/_bulk", bw.w)
	if err != nil {
		log.Fatalf("Couldn't make a request: %v\n", err)
	}
	req.Header.Set("Content-Length", fmt.Sprintf("%d", bw.w.Len()))
	req.Header.Set("Content-Type", "application/json")
	reqch <- req
	bw.w = &bytes.Buffer{}
}

// Get a bulk updater.
func (e *ElasticSearch) Bulk() BulkUpdater {

	rv := &bulkWriter{
		update: make(chan Instruction),
		reqch:  make(chan chan *http.Request),
		quit:   make(chan bool),
		w:      &bytes.Buffer{},
	}

	go func() {
		ever := true
		for ever {
			select {
			case <-rv.quit:
				ever = false
			case req := <-rv.reqch:
				issueBulkRequest(e.URL, rv, req)
			case upd := <-rv.update:
				err := upd.writeTo(rv.w)
				if err != nil {
					log.Fatalf("Error sending an update: %v", err)
				}
			}
		}
	}()

	return rv
}
