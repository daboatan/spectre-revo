package main

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/DHowett/gotimeout"
	"github.com/golang/glog"
)

type GrantStore struct {
	Grants     map[GrantID]PasteID
	ExpiryJunk *gotimeout.HandleMap

	filename  string
	expirator *gotimeout.Expirator
	mu        sync.RWMutex
}

type GrantID string

func (r *GrantStore) saveLocked() error {
	asideFilename := r.filename + ".atomic"
	file, err := os.Create(asideFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := gob.NewEncoder(file)

	err = enc.Encode(r)
	if err != nil {
		glog.Error("Failed to save grants: ", err)
		return err
	}

	return os.Rename(asideFilename, r.filename)
}

func (r *GrantStore) NewGrant(id PasteID) GrantID {
	newKey, _ := generateRandomBase32String(20, 32)
	grantKey := GrantID(newKey)
	r.mu.Lock()
	r.Grants[grantKey] = id
	r.mu.Unlock()

	// The expirator persists the grant and its expiration together on its
	// urgent flush cycle.
	r.expirator.ExpireObject(grantKey, 48*time.Hour)

	return grantKey
	//GrantID returned
}

func (r *GrantStore) Delete(p GrantID) {
	r.expirator.CancelObjectExpiration(p)
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Grants, p)
}

func (r *GrantStore) Get(p GrantID) (PasteID, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pid, ok := r.Grants[p]
	return pid, ok
}

func LoadGrantStore(filename string) *GrantStore {
	var gs *GrantStore
	grant_file, err := os.Open(filename)
	if err == nil {
		dec := gob.NewDecoder(grant_file)
		err := dec.Decode(&gs)

		if err != nil {
			glog.Error("Failed to decode grants: ", err)
		}
	}
	if gs == nil {
		gs = &GrantStore{}
	}
	if gs.Grants == nil {
		gs.Grants = make(map[GrantID]PasteID)
	}
	gs.filename = filename
	gs.expirator = gotimeout.NewExpiratorWithStorage(gs, gs)
	return gs
}

// grantKey passed in and made sure it exists.
func (e *GrantStore) GetExpirable(grantKey gotimeout.ExpirableID) gotimeout.Expirable {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.Grants[GrantID(grantKey)]
	if !ok {
		return nil
	}
	return GrantID(grantKey)
}

func (e *GrantStore) DestroyExpirable(ex gotimeout.Expirable) {
	if grant, ok := ex.(GrantID); ok {
		e.Delete(grant)
	}
}

func (p GrantID) ExpirationID() gotimeout.ExpirableID {
	return gotimeout.ExpirableID(p)
}

func (e *GrantStore) RequiresFlush() bool {
	return true
}

func (e *GrantStore) SaveExpirationHandles(hm *gotimeout.HandleMap) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ExpiryJunk = hm
	return e.saveLocked()
}

func (e *GrantStore) LoadExpirationHandles() (*gotimeout.HandleMap, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.ExpiryJunk, nil
}

var grantStore *GrantStore

func init() {
	arguments.register()
	arguments.parse()
	grantStore = LoadGrantStore(filepath.Join(arguments.root, "grants.gob"))
}
