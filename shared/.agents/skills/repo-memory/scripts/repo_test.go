package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestIdentity(t *testing.T) {
	for _, remote := range []string{
		"git@github.com:org/repo.git",
		"https://user:secret@github.com/org/repo.git",
		"ssh://git@github.com:22/org/repo.git",
	} {
		got, err := identity(remote, t.TempDir())
		if err != nil || got != "github.com/org/repo" {
			t.Fatalf("identity = %q, %v", got, err)
		}
	}
	root := t.TempDir()
	if got, err := identity("", root); err != nil || got != "path:"+root {
		t.Fatalf("fallback = %q, %v", got, err)
	}
	if got, err := identity("https://example.com:8443/org/Repo", root); err != nil || got != "example.com:8443/org/Repo" {
		t.Fatalf("port = %q, %v", got, err)
	}
	if _, err := identity("file:///tmp/repo", root); err == nil {
		t.Fatal("unsupported URL accepted")
	}
}

func TestMailbox(t *testing.T) {
	// Build a disposable Git fixture without running mutating Git commands.
	root, shared := t.TempDir(), t.TempDir()
	for _, dir := range []string{"objects", "refs"} {
		if err := os.MkdirAll(filepath.Join(root, ".git", dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	setOrigin := func(remote string) {
		t.Helper()
		config := "[core]\nrepositoryformatversion = 0\nbare = false\n[remote \"origin\"]\nurl = " + remote + "\n"
		if err := os.WriteFile(filepath.Join(root, ".git", "config"), []byte(config), 0600); err != nil {
			t.Fatal(err)
		}
	}
	setOrigin("https://example.com/org/frontend.git")
	invoke := func(args ...string) (map[string]any, error) {
		var out bytes.Buffer
		err := run(append([]string{"--repo", root}, args...), shared, &out)
		if err != nil {
			return nil, err
		}
		var result map[string]any
		err = json.Unmarshal(out.Bytes(), &result)
		return result, err
	}
	must := func(args ...string) map[string]any {
		t.Helper()
		result, err := invoke(args...)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	status := func() string {
		t.Helper()
		return must("list")["messages"].([]any)[0].(map[string]any)["status"].(string)
	}
	draft := filepath.Join(t.TempDir(), "draft.md")
	if err := os.WriteFile(draft, []byte("# Frontend task\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if len(must("list")["messages"].([]any)) != 0 {
		t.Fatal("mailbox is not empty")
	}
	msg := must("send", "--file", draft)
	id := msg["id"].(string)
	if status() != "pending" {
		t.Fatal("expected pending")
	}
	if _, err := invoke("claim", "../../outside"); err == nil {
		t.Fatal("unsafe ID accepted")
	}
	claim := must("claim", id)
	if status() != "claimed" {
		t.Fatal("expected claimed")
	}
	if _, err := invoke("claim", id); err == nil {
		t.Fatal("duplicate claim accepted")
	}
	if _, err := invoke("finish", id, "--token", "wrong", "--status", "done", "--file", draft); err == nil {
		t.Fatal("wrong token accepted")
	}
	token := claim["token"].(string)
	must("finish", id, "--token", token, "--status", "done", "--file", draft)
	if status() != "done" {
		t.Fatal("expected done")
	}
	if _, err := invoke("finish", id, "--token", token, "--status", "done", "--file", draft); err == nil {
		t.Fatal("duplicate finish accepted")
	}

	blocked := must("send", "--file", draft)["id"].(string)
	blockedToken := must("claim", blocked)["token"].(string)
	must("finish", blocked, "--token", blockedToken, "--status", "blocked", "--file", draft)
	var result outcome
	loc, err := locate(root, shared)
	if err != nil {
		t.Fatal(err)
	}
	inbox := filepath.Join(shared, "handoffs", filepath.Base(loc.Directory))
	if err := readJSON(filepath.Join(inbox, blocked, "outcome.json"), &result); err != nil || result.Status != "blocked" {
		t.Fatalf("blocked outcome: %+v, %v", result, err)
	}

	concurrent := must("send", "--file", draft)["id"].(string)
	var winners atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := invoke("claim", concurrent); err == nil {
				winners.Add(1)
			}
		}()
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("claim winners = %d", winners.Load())
	}
	setOrigin("https://example.com/org/other.git")
	if len(must("list")["messages"].([]any)) != 0 {
		t.Fatal("mail crossed repo identities")
	}
	if err := os.WriteFile(draft, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := invoke("send", "--file", draft); err == nil {
		t.Fatal("empty draft accepted")
	}
}

func TestMissingMount(t *testing.T) {
	if _, err := locate(t.TempDir(), filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("missing mount accepted")
	}
}
