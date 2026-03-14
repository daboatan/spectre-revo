package main

import (
	"net/http"

	"github.com/borrougagnou/spectre-updated/account"
	"github.com/golang/glog"
)

type PastePermission map[string]bool

type PastePermissionSet struct {
	Entries map[PasteID]PastePermission
	u       *account.User
}

func GetPastePermissions(r *http.Request) *PastePermissionSet {
	var perms *PastePermissionSet

	// Check if we have a user first.
	user := GetUser(r)
	if user != nil {
		if userPerms, ok := user.Values["permissions"]; ok {
			if p, ok := userPerms.(*PastePermissionSet); ok {
				perms = p
			} else if p, ok := userPerms.(PastePermissionSet); ok {
				perms = &p
			}
		}
		if perms != nil {
			glog.Infof("Found existing permissions for user %s (%d entries)", user.Name, len(perms.Entries))
		}
	}

	cookieSession, _ := sessionStore.Get(r, "session")

	// Attempt to get hold of the new-style permission set.
	if sessionPermissionSet, ok := cookieSession.Values["permissions"]; ok {
		var sessionPerms *PastePermissionSet
		if p, ok := sessionPermissionSet.(*PastePermissionSet); ok {
			sessionPerms = p
		} else if p, ok := sessionPermissionSet.(PastePermissionSet); ok {
			sessionPerms = &p
		}

		if sessionPerms != nil && len(sessionPerms.Entries) > 0 {
			glog.Infof("Merging session permissions (%d entries) for user %v", len(sessionPerms.Entries), user != nil)
			if perms == nil {
				perms = sessionPerms
			} else {
				if perms.Entries == nil {
					perms.Entries = make(map[PasteID]PastePermission)
				}
				for k, v := range sessionPerms.Entries {
					perms.Put(k, v)
				}
			}
		}
	}

	if perms == nil {
		perms = &PastePermissionSet{
			Entries: make(map[PasteID]PastePermission),
		}
	}

	if perms.Entries == nil {
		perms.Entries = make(map[PasteID]PastePermission)
	}

	if user != nil {
		user.Values["permissions"] = perms
	}

	// Attempt to get hold of the original list of pastes
	if oldPasteList, ok := cookieSession.Values["pastes"]; ok {
		if pastes, ok := oldPasteList.([]string); ok {
			glog.Infof("Merging legacy pastes (%d entries)", len(pastes))
			for _, v := range pastes {
				perms.Put(PasteIDFromString(v), PastePermission{
					"grant": true,
					"edit":  true,
				})
			}
		}
	}

	perms.u = user
	return perms
}

// Save emits the PastePermissionSet to disk, either as part of the anonymous
// session or as part of the authenticated user's data.
func (p *PastePermissionSet) Save(w http.ResponseWriter, r *http.Request) {
	if p.u != nil {
		p.u.Save()
	} else {
		cookieSession, _ := sessionStore.Get(r, "session")
		cookieSession.Values["permissions"] = p
		cookieSession.Save(r, w)
	}
}

// Put inserts a set of permissions into the permission store,
// potentially merging new permissions with existing permissions for the same paste.
func (p *PastePermissionSet) Put(id PasteID, perms PastePermission) {
	if existing, ok := p.Entries[id]; ok {
		for k, v := range perms {
			existing[k] = v
		}
	} else {
		p.Entries[id] = perms
	}
}

func (p *PastePermissionSet) Get(id PasteID) (PastePermission, bool) {
	v, ok := p.Entries[id]
	return v, ok
}

func (p *PastePermissionSet) Delete(id PasteID) {
	delete(p.Entries, id)
}
