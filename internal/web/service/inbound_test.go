package service

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

// initInboundTestDB initialises a throw-away SQLite database in t.TempDir()
// and registers a cleanup that closes it. Returns the db folder path so
// callers can derive it if needed; most tests just use database.GetDB().
// Mirrors the pattern from internal/database/db_seed_test.go.
func initInboundTestDB(t *testing.T) string {
	t.Helper()
	dbDir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dbDir)
	if err := database.InitDB(filepath.Join(dbDir, "x-ui.db")); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	t.Cleanup(func() {
		if err := database.CloseDB(); err != nil {
			t.Logf("CloseDB: %v", err)
		}
	})
	return dbDir
}

// seedTestInbound inserts a minimal VLESS inbound for testing and returns it.
func seedTestInbound(t *testing.T, overrides map[string]any) *model.Inbound {
	t.Helper()
	db := database.GetDB()

	clients := []any{
		map[string]any{
			"id":     "ce8d33df-3a64-4f10-8f9b-91c3a8e0c001",
			"email":  "alice@example.com",
			"enable": true,
			"subId":  "alice-sub",
			"flow":   "",
		},
		map[string]any{
			"id":     "ce8d33df-3a64-4f10-8f9b-91c3a8e0c002",
			"email":  "bob@example.com",
			"enable": true,
			"subId":  "bob-sub",
			"flow":   "xtls-rprx-vision",
		},
	}
	settings, err := json.Marshal(map[string]any{"clients": clients})
	if err != nil {
		t.Fatalf("marshal settings: %v", err)
	}

	inbound := model.Inbound{
		UserId:   1,
		Port:     443,
		Protocol: model.VLESS,
		Settings: string(settings),
		Tag:      "test-inbound",
		Remark:   "Test VLESS",
		Enable:   true,
	}

	// Apply any overrides after defaults so callers can tweak fields.
	if overrides != nil {
		if v, ok := overrides["port"]; ok {
			inbound.Port = v.(int)
		}
		if v, ok := overrides["protocol"]; ok {
			inbound.Protocol = model.Protocol(v.(string))
		}
		if v, ok := overrides["tag"]; ok {
			inbound.Tag = v.(string)
		}
		if v, ok := overrides["remark"]; ok {
			inbound.Remark = v.(string)
		}
		if v, ok := overrides["enable"]; ok {
			inbound.Enable = v.(bool)
		}
	}

	if err := db.Create(&inbound).Error; err != nil {
		t.Fatalf("seed inbound: %v", err)
	}
	return &inbound
}

// TestInboundService_GetInbounds verifies that GetInbounds returns all
// inbounds for a user, preloaded with ClientStats.
func TestInboundService_GetInbounds(t *testing.T) {
	initInboundTestDB(t)
	ib := seedTestInbound(t, nil)

	svc := InboundService{}
	inbounds, err := svc.GetInbounds(1)
	if err != nil {
		t.Fatalf("GetInbounds: %v", err)
	}
	if len(inbounds) != 1 {
		t.Fatalf("expected 1 inbound, got %d", len(inbounds))
	}
	if inbounds[0].Id != ib.Id {
		t.Errorf("inbound.Id = %d, want %d", inbounds[0].Id, ib.Id)
	}
	if inbounds[0].Port != 443 {
		t.Errorf("inbound.Port = %d, want 443", inbounds[0].Port)
	}
	if inbounds[0].Protocol != model.VLESS {
		t.Errorf("inbound.Protocol = %q, want %q", inbounds[0].Protocol, model.VLESS)
	}
}

// TestInboundService_GetInbounds_emptyUser verifies GetInbounds returns an
// empty slice (not nil) when the user has no inbounds.
func TestInboundService_GetInbounds_emptyUser(t *testing.T) {
	initInboundTestDB(t)
	// Seed one inbound for user 1, then query user 2.
	seedTestInbound(t, nil)

	svc := InboundService{}
	inbounds, err := svc.GetInbounds(2)
	if err != nil {
		t.Fatalf("GetInbounds: %v", err)
	}
	if len(inbounds) != 0 {
		t.Fatalf("expected 0 inbounds for user 2, got %d", len(inbounds))
	}
}

// TestInboundService_GetInbound verifies fetching a single inbound by id.
func TestInboundService_GetInbound(t *testing.T) {
	initInboundTestDB(t)
	ib := seedTestInbound(t, nil)

	svc := InboundService{}
	got, err := svc.GetInbound(ib.Id)
	if err != nil {
		t.Fatalf("GetInbound(%d): %v", ib.Id, err)
	}
	if got.Id != ib.Id {
		t.Errorf("got.Id = %d, want %d", got.Id, ib.Id)
	}
	if got.Tag != "test-inbound" {
		t.Errorf("got.Tag = %q, want %q", got.Tag, "test-inbound")
	}
}

// TestInboundService_GetInbound_notFound verifies GetInbound returns an error
// for a non-existent id.
func TestInboundService_GetInbound_notFound(t *testing.T) {
	initInboundTestDB(t)

	svc := InboundService{}
	_, err := svc.GetInbound(99999)
	if err == nil {
		t.Fatal("expected error for non-existent inbound, got nil")
	}
}

// TestInboundService_GetInbounds_sameUserMultipleInbounds verifies that a
// user with multiple inbounds gets them all back, ordered by id ASC.
func TestInboundService_GetInbounds_sameUserMultipleInbounds(t *testing.T) {
	initInboundTestDB(t)
	ib1 := seedTestInbound(t, map[string]any{"port": 443, "tag": "in-443"})
	ib2 := seedTestInbound(t, map[string]any{"port": 8443, "tag": "in-8443"})
	ib3 := seedTestInbound(t, map[string]any{"port": 2053, "tag": "in-2053"})

	svc := InboundService{}
	inbounds, err := svc.GetInbounds(1)
	if err != nil {
		t.Fatalf("GetInbounds: %v", err)
	}
	if len(inbounds) != 3 {
		t.Fatalf("expected 3 inbounds, got %d", len(inbounds))
	}
	// Verify order: ids should be ascending.
	ids := []int{ib1.Id, ib2.Id, ib3.Id}
	for i, got := range inbounds {
		if got.Id != ids[i] {
			t.Errorf("inbounds[%d].Id = %d, want %d", i, got.Id, ids[i])
		}
	}
}

// TestInboundService_GetAllInbounds verifies GetAllInbounds returns every
// inbound across all users.
func TestInboundService_GetAllInbounds(t *testing.T) {
	initInboundTestDB(t)
	seedTestInbound(t, nil)

	svc := InboundService{}
	inbounds, err := svc.GetAllInbounds()
	if err != nil {
		t.Fatalf("GetAllInbounds: %v", err)
	}
	if len(inbounds) < 1 {
		t.Fatal("expected at least 1 inbound")
	}
}

// TestInboundService_GetClients verifies client extraction from an inbound's
// settings JSON.
func TestInboundService_GetClients(t *testing.T) {
	initInboundTestDB(t)
	ib := seedTestInbound(t, nil)

	svc := InboundService{}
	clients, err := svc.GetClients(ib)
	if err != nil {
		t.Fatalf("GetClients: %v", err)
	}
	if len(clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(clients))
	}
	if clients[0].Email != "alice@example.com" {
		t.Errorf("clients[0].Email = %q, want %q", clients[0].Email, "alice@example.com")
	}
	if clients[1].Email != "bob@example.com" {
		t.Errorf("clients[1].Email = %q, want %q", clients[1].Email, "bob@example.com")
	}
}

// TestInboundService_GetClients_noClients verifies that an inbound with an
// empty clients array returns nil without error.
func TestInboundService_GetClients_noClients(t *testing.T) {
	initInboundTestDB(t)
	settings, err := json.Marshal(map[string]any{"clients": []any{}})
	if err != nil {
		t.Fatalf("marshal settings: %v", err)
	}
	ib := seedTestInbound(t, nil)
	ib.Settings = string(settings)

	svc := InboundService{}
	clients, err := svc.GetClients(ib)
	if err != nil {
		t.Fatalf("GetClients: %v", err)
	}
	if len(clients) != 0 {
		t.Fatalf("expected 0 clients, got %d", len(clients))
	}
}

// TestInboundService_GetInboundOptions verifies the slim options list has the
// expected fields for each inbound.
func TestInboundService_GetInboundOptions(t *testing.T) {
	initInboundTestDB(t)
	seedTestInbound(t, nil)

	svc := InboundService{}
	opts, err := svc.GetInboundOptions(1)
	if err != nil {
		t.Fatalf("GetInboundOptions: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("expected 1 option, got %d", len(opts))
	}
	o := opts[0]
	if o.Id == 0 {
		t.Errorf("option.Id is 0, expected non-zero")
	}
	if o.Remark != "Test VLESS" {
		t.Errorf("option.Remark = %q, want %q", o.Remark, "Test VLESS")
	}
	if o.Protocol != "vless" {
		t.Errorf("option.Protocol = %q, want %q", o.Protocol, "vless")
	}
	if o.Port != 443 {
		t.Errorf("option.Port = %d, want %d", o.Port, 443)
	}
}

// TestInboundService_GetAllEmails verifies that GetAllEmails returns unique
// email addresses across all inbounds.
func TestInboundService_GetAllEmails(t *testing.T) {
	initInboundTestDB(t)
	seedTestInbound(t, nil)

	svc := InboundService{}
	emails, err := svc.GetAllEmails()
	if err != nil {
		t.Fatalf("GetAllEmails: %v", err)
	}
	// Should have both alice and bob.
	seen := make(map[string]bool)
	for _, e := range emails {
		seen[e] = true
	}
	if !seen["alice@example.com"] {
		t.Errorf("missing alice@example.com in emails")
	}
	if !seen["bob@example.com"] {
		t.Errorf("missing bob@example.com in emails")
	}
}

// TestInboundService_GetInboundsByTrafficReset verifies the traffic reset
// period filter works.
func TestInboundService_GetInboundsByTrafficReset(t *testing.T) {
	initInboundTestDB(t)
	ib := seedTestInbound(t, nil)
	ib.TrafficReset = "daily"
	db := database.GetDB()
	if err := db.Save(ib).Error; err != nil {
		t.Fatalf("update inbound traffic_reset: %v", err)
	}

	svc := InboundService{}
	inbounds, err := svc.GetInboundsByTrafficReset("daily")
	if err != nil {
		t.Fatalf("GetInboundsByTrafficReset: %v", err)
	}
	if len(inbounds) != 1 {
		t.Fatalf("expected 1 inbound with daily reset, got %d", len(inbounds))
	}
}

// TestInboundService_GetInboundsByTrafficReset_noMatch verifies that
// querying for an unused reset period returns an empty slice.
func TestInboundService_GetInboundsByTrafficReset_noMatch(t *testing.T) {
	initInboundTestDB(t)
	seedTestInbound(t, nil)

	svc := InboundService{}
	inbounds, err := svc.GetInboundsByTrafficReset("monthly")
	if err != nil {
		t.Fatalf("GetInboundsByTrafficReset: %v", err)
	}
	if len(inbounds) != 0 {
		t.Fatalf("expected 0 inbounds, got %d", len(inbounds))
	}
}

// TestInboundService_GetInboundTags verifies the tags JSON string.
func TestInboundService_GetInboundTags(t *testing.T) {
	initInboundTestDB(t)
	seedTestInbound(t, nil)

	svc := InboundService{}
	tagsJSON, err := svc.GetInboundTags()
	if err != nil {
		t.Fatalf("GetInboundTags: %v", err)
	}
	if tagsJSON == "" || tagsJSON == "[]" {
		t.Fatalf("expected non-empty tags JSON, got %q", tagsJSON)
	}
}

// TestInboundService_SearchInbounds verifies searching by remark fragment.
func TestInboundService_SearchInbounds(t *testing.T) {
	initInboundTestDB(t)
	seedTestInbound(t, nil)

	svc := InboundService{}
	inbounds, err := svc.SearchInbounds("Test")
	if err != nil {
		t.Fatalf("SearchInbounds: %v", err)
	}
	if len(inbounds) != 1 {
		t.Fatalf("expected 1 inbound matching 'Test', got %d", len(inbounds))
	}
}

// TestInboundService_SearchInbounds_noMatch verifies no results for a
// non-matching query.
func TestInboundService_SearchInbounds_noMatch(t *testing.T) {
	initInboundTestDB(t)
	seedTestInbound(t, nil)

	svc := InboundService{}
	inbounds, err := svc.SearchInbounds("NonExistent")
	if err != nil {
		t.Fatalf("SearchInbounds: %v", err)
	}
	if len(inbounds) != 0 {
		t.Fatalf("expected 0 results, got %d", len(inbounds))
	}
}

// TestInboundService_GetInboundsSlim verifies the slim list drops client
// detail fields but preserves email/enable/comment.
func TestInboundService_GetInboundsSlim(t *testing.T) {
	initInboundTestDB(t)
	seedTestInbound(t, nil)

	svc := InboundService{}
	inbounds, err := svc.GetInboundsSlim(1)
	if err != nil {
		t.Fatalf("GetInboundsSlim: %v", err)
	}
	if len(inbounds) != 1 {
		t.Fatalf("expected 1 inbound, got %d", len(inbounds))
	}
	// The slim settings should still contain client email/enable but drop
	// heavy fields like the UUID in "id".
	ib := inbounds[0]
	if ib.Settings == "" {
		t.Fatal("slim settings should not be empty")
	}
	// Verify the IDs are stripped from the slim output.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(ib.Settings), &parsed); err != nil {
		t.Fatalf("unmarshal slim settings: %v", err)
	}
	clientsRaw, ok := parsed["clients"].([]any)
	if !ok {
		t.Fatal("slim settings missing clients array")
	}
	for _, raw := range clientsRaw {
		c, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if _, hasID := c["id"]; hasID {
			t.Errorf("slim client should not carry id, got %v", c["id"])
		}
		if _, hasFlow := c["flow"]; hasFlow {
			t.Errorf("slim client should not carry flow, got %v", c["flow"])
		}
		if email, hasEmail := c["email"]; !hasEmail || email == "" {
			t.Errorf("slim client should carry email, got %v", email)
		}
	}
}
