package main

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

type ReportInfo map[string]int
type ReportStore struct {
	Reports  map[PasteID]ReportInfo
	filename string
	mu       sync.RWMutex
}

func (r *ReportStore) Save() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saveLocked()
}

func (r *ReportStore) saveLocked() error {
	asideFilename := r.filename + ".atomic"
	file, err := os.Create(asideFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := gob.NewEncoder(file)

	err = enc.Encode(r)
	if err != nil {
		glog.Error("Failed to save reports: ", err)
		return err
	}

	return os.Rename(asideFilename, r.filename)
}

func (r *ReportStore) Add(id PasteID, kind string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentReportsForPaste, ok := r.Reports[id]

	if !ok {
		currentReportsForPaste = make(ReportInfo)
		r.Reports[id] = currentReportsForPaste
	}

	currentReportsForPaste[kind] = currentReportsForPaste[kind] + 1
	r.saveLocked()
}

func (r *ReportStore) Delete(p PasteID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Reports, p)
	glog.Info(p, " deleted from report history.")
	r.saveLocked()
}

func (r *ReportStore) Snapshot() map[PasteID]ReportInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reports := make(map[PasteID]ReportInfo, len(r.Reports))
	for id, info := range r.Reports {
		clone := make(ReportInfo, len(info))
		for kind, count := range info {
			clone[kind] = count
		}
		reports[id] = clone
	}
	return reports
}

func LoadReportStore(filename string) *ReportStore {
	report_file, err := os.Open(filename)
	if err == nil {
		var decoded_reports *ReportStore
		dec := gob.NewDecoder(report_file)
		err := dec.Decode(&decoded_reports)

		if err == nil {
			decoded_reports.filename = filename
			return decoded_reports
		} else {
			glog.Error("Failed to decode reports: ", err)
		}
	}
	return &ReportStore{Reports: map[PasteID]ReportInfo{}, filename: filename}
}

var reportStore *ReportStore

func reportPaste(o Model, w http.ResponseWriter, r *http.Request) {
	if throttleAuthForRequest(r) {
		RenderError(fmt.Errorf("Cool it."), 420, w)
		return
	}

	p := o.(*Paste)
	reason := r.FormValue("reason")

	reportStore.Add(p.ID, reason)

	SetFlash(w, "success", fmt.Sprintf("Paste %v reported.", p.ID))
	w.Header().Set("Location", pasteURL("show", p))
	w.WriteHeader(http.StatusFound)
}

func reportClear(w http.ResponseWriter, r *http.Request) {
	defer errorRecoveryHandler(w)

	id := PasteIDFromString(mux.Vars(r)["id"])
	reportStore.Delete(id)

	SetFlash(w, "success", fmt.Sprintf("Report for %v cleared.", id))
	w.Header().Set("Location", "/admin/reports")
	w.WriteHeader(http.StatusFound)
}

func init() {
	arguments.register()
	arguments.parse()
	reportStore = LoadReportStore(filepath.Join(arguments.root, "reports.gob"))
}
