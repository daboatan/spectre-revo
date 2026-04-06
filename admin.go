package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/golang/glog"
)

type AdminDashboardSnapshot struct {
	TotalPastes    int
	TotalReports   int
	TotalUsers     int
	TotalAdmins    int
	ExpiringPastes int
}

type AdminReportRow struct {
	PasteID      PasteID
	Reasons      []string
	TotalReports int
	Preview      string
}

type AdminPasteRow struct {
	PasteID       PasteID
	Title         string
	Language      string
	LastModified  time.Time
	ExpiresAt     *time.Time
	IsEncrypted   bool
	ReportCount   int
	ReportReasons []string
}

type AdminUserRow struct {
	Name      string
	IsAdmin   bool
	IsPersona bool
}

type AdminHomeView struct {
	Dashboard AdminDashboardSnapshot
}

type AdminDashboardView struct {
	Dashboard     AdminDashboardSnapshot
	RecentReports []AdminReportRow
}

type AdminReportsView struct {
	Dashboard AdminDashboardSnapshot
	Reports   []AdminReportRow
}

type AdminPastesView struct {
	Dashboard    AdminDashboardSnapshot
	Pastes       []AdminPasteRow
	Query        string
	ReportedOnly bool
}

type AdminUsersView struct {
	Dashboard AdminDashboardSnapshot
	Users     []AdminUserRow
	Query     string
	AdminOnly bool
}

func adminAccountPath() string {
	return filepath.Join(arguments.root, "accounts")
}

func adminPastePath() string {
	return filepath.Join(arguments.root, "pastes")
}

func listAccountNames() []string {
	entries, err := os.ReadDir(adminAccountPath())
	if err != nil {
		glog.Error("failed to read account directory: ", err)
		return nil
	}

	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".tmp") {
			continue
		}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func listPasteIDs() []PasteID {
	entries, err := os.ReadDir(adminPastePath())
	if err != nil {
		glog.Error("failed to read paste directory: ", err)
		return nil
	}

	out := make([]PasteID, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		out = append(out, PasteIDFromString(e.Name()))
	}
	return out
}

func summarizeReportInfo(info ReportInfo) ([]string, int) {
	reasons := make([]string, 0, len(info))
	total := 0
	for reason, count := range info {
		reasons = append(reasons, fmt.Sprintf("%s x%d", reason, count))
		total += count
	}
	sort.Strings(reasons)
	return reasons, total
}

func pastePreview(p *Paste, maxLines int) string {
	if p == nil {
		return ""
	}
	reader, err := p.Reader()
	if err != nil {
		return ""
	}
	defer reader.Close()

	buf := make([]byte, 0, 512)
	tmp := make([]byte, 256)
	lineCount := 0
	for lineCount < maxLines {
		n, readErr := reader.Read(tmp)
		if n > 0 {
			chunk := tmp[:n]
			for i := 0; i < len(chunk); i++ {
				if chunk[i] == '\n' {
					lineCount++
					if lineCount >= maxLines {
						buf = append(buf, chunk[:i]...)
						if readErr == nil {
							return string(buf) + "..."
						}
						return string(buf)
					}
				}
			}
			buf = append(buf, chunk...)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			break
		}
	}
	return string(buf)
}

func loadPasteForAdmin(id PasteID) (*Paste, bool) {
	p, err := pasteStore.Get(id, nil)
	if err == nil {
		return p, true
	}

	if _, ok := err.(PasteEncryptedError); ok && p != nil {
		return p, true
	}

	return nil, false
}

func buildAdminDashboardSnapshot() AdminDashboardSnapshot {
	accountNames := listAccountNames()
	pasteIDs := listPasteIDs()
	reports := reportStore.Snapshot()

	adminCount := 0
	for _, name := range accountNames {
		u := userStore.Get(name)
		if u == nil {
			continue
		}
		if perms, ok := u.Values["user.permissions"].(PastePermission); ok && perms["admin"] {
			adminCount++
		}
	}

	reportTotal := 0
	for _, info := range reports {
		for _, count := range info {
			reportTotal += count
		}
	}

	return AdminDashboardSnapshot{
		TotalPastes:    len(pasteIDs),
		TotalReports:   reportTotal,
		TotalUsers:     len(accountNames),
		TotalAdmins:    adminCount,
		ExpiringPastes: pasteExpirator.Len(),
	}
}

func buildAdminReportsRows() []AdminReportRow {
	reports := reportStore.Snapshot()
	rows := make([]AdminReportRow, 0, len(reports))
	for pasteID, info := range reports {
		reasons, total := summarizeReportInfo(info)
		row := AdminReportRow{
			PasteID:      pasteID,
			Reasons:      reasons,
			TotalReports: total,
		}
		if p, ok := loadPasteForAdmin(pasteID); ok {
			if p.Encrypted {
				row.Preview = "(encrypted paste)"
			} else {
				row.Preview = pastePreview(p, 5)
			}
		} else {
			row.Preview = "(paste unavailable)"
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].TotalReports == rows[j].TotalReports {
			return rows[i].PasteID.String() < rows[j].PasteID.String()
		}
		return rows[i].TotalReports > rows[j].TotalReports
	})
	return rows
}

func buildAdminPasteRows(query string, reportedOnly bool) []AdminPasteRow {
	pasteIDs := listPasteIDs()
	reports := reportStore.Snapshot()
	q := strings.ToLower(strings.TrimSpace(query))

	rows := make([]AdminPasteRow, 0, len(pasteIDs))
	for _, id := range pasteIDs {
		p, ok := loadPasteForAdmin(id)
		if !ok || p == nil {
			continue
		}

		info := reports[id]
		reasons, reportCount := summarizeReportInfo(info)
		if reportedOnly && reportCount == 0 {
			continue
		}

		row := AdminPasteRow{
			PasteID:       id,
			Title:         p.Title,
			Language:      p.Language.ID,
			LastModified:  p.LastModified(),
			IsEncrypted:   p.Encrypted,
			ReportCount:   reportCount,
			ReportReasons: reasons,
		}

		if pasteWillExpire := p.Expiration != "" && p.Expiration != "-1"; pasteWillExpire {
			expiresAt := p.ExpirationTime()
			row.ExpiresAt = &expiresAt
		}

		if q != "" {
			hay := strings.ToLower(strings.Join([]string{
				row.PasteID.String(),
				row.Title,
				row.Language,
				strings.Join(row.ReportReasons, " "),
			}, " "))
			if !strings.Contains(hay, q) {
				continue
			}
		}

		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].LastModified.After(rows[j].LastModified)
	})

	return rows
}

func buildAdminUserRows(query string, adminOnly bool) []AdminUserRow {
	names := listAccountNames()
	q := strings.ToLower(strings.TrimSpace(query))

	rows := make([]AdminUserRow, 0, len(names))
	for _, name := range names {
		u := userStore.Get(name)
		if u == nil {
			continue
		}

		isAdmin := false
		if perms, ok := u.Values["user.permissions"].(PastePermission); ok && perms["admin"] {
			isAdmin = true
		}
		if adminOnly && !isAdmin {
			continue
		}

		isPersona, _ := u.Values["persona"].(bool)
		row := AdminUserRow{
			Name:      u.Name,
			IsAdmin:   isAdmin,
			IsPersona: isPersona,
		}

		if q != "" && !strings.Contains(strings.ToLower(row.Name), q) {
			continue
		}

		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].IsAdmin == rows[j].IsAdmin {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].IsAdmin && !rows[j].IsAdmin
	})
	return rows
}

func adminHomeHandler(w http.ResponseWriter, r *http.Request) {
	RenderPage(w, r, "admin_home", AdminHomeView{
		Dashboard: buildAdminDashboardSnapshot(),
	})
}

func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	reports := buildAdminReportsRows()
	if len(reports) > 5 {
		reports = reports[:5]
	}
	RenderPage(w, r, "admin_dashboard", AdminDashboardView{
		Dashboard:     buildAdminDashboardSnapshot(),
		RecentReports: reports,
	})
}

func adminReportsHandler(w http.ResponseWriter, r *http.Request) {
	RenderPage(w, r, "admin_reports", AdminReportsView{
		Dashboard: buildAdminDashboardSnapshot(),
		Reports:   buildAdminReportsRows(),
	})
}

func adminPastesHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	reportedOnly := r.URL.Query().Get("reported") == "1"
	RenderPage(w, r, "admin_pastes", AdminPastesView{
		Dashboard:    buildAdminDashboardSnapshot(),
		Pastes:       buildAdminPasteRows(query, reportedOnly),
		Query:        query,
		ReportedOnly: reportedOnly,
	})
}

func adminUsersHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	adminOnly := r.URL.Query().Get("admin") == "1"
	RenderPage(w, r, "admin_users", AdminUsersView{
		Dashboard: buildAdminDashboardSnapshot(),
		Users:     buildAdminUserRows(query, adminOnly),
		Query:     query,
		AdminOnly: adminOnly,
	})
}
