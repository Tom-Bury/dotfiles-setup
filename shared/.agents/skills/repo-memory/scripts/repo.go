// Run with: go run /absolute/path/to/repo.go [--repo checkout] <command>.
// Commands: location, send --file draft, list, claim ID,
// finish ID --token TOKEN --status done|blocked --file report.
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type location struct {
	Repo         string `json:"repo"`
	PathFallback bool   `json:"path_fallback"`
	Directory    string `json:"directory"`
}

type message struct {
	ID        string `json:"id"`
	Repo      string `json:"repo"`
	Created   string `json:"created"`
	Status    string `json:"status,omitempty"`
	Directory string `json:"directory,omitempty"`
}

type ownership struct {
	Token   string `json:"token"`
	Claimed string `json:"claimed"`
}

type outcome struct {
	Status   string `json:"status"`
	Finished string `json:"finished"`
}

func git(repo string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	data, err := cmd.Output()
	return strings.TrimSpace(string(data)), err
}

var scpRemote = regexp.MustCompile(`^(?:[^/@:]+@)?([^/:]+):(.+)$`)
var validID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func identity(remote, root string) (string, error) {
	local := func(path string) (string, error) {
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			resolved, err = filepath.Abs(path)
		}
		return "path:" + resolved, err
	}
	if remote == "" {
		return local(root)
	}
	var host, path string
	if strings.Contains(remote, "://") {
		u, err := url.Parse(remote)
		if err != nil {
			return "", errors.New("invalid origin URL")
		}
		defaults := map[string]string{"https": "443", "http": "80", "ssh": "22", "git": "9418"}
		port, supported := defaults[u.Scheme]
		if !supported || u.Hostname() == "" {
			return "", errors.New("unsupported origin URL")
		}
		host = strings.ToLower(u.Hostname())
		if u.Port() != "" && u.Port() != port {
			host = net.JoinHostPort(host, u.Port())
		}
		path = u.Path
	} else if match := scpRemote.FindStringSubmatch(remote); match != nil {
		host, path = strings.ToLower(match[1]), match[2]
	} else {
		return local(remote)
	}
	path = strings.TrimSuffix(strings.Trim(path, "/"), ".git")
	if path == "" {
		return "", errors.New("origin URL has no repo path")
	}
	return host + "/" + path, nil
}

func locate(repo, shared string) (location, error) {
	info, err := os.Stat(shared)
	if err != nil || !info.IsDir() {
		return location{}, errors.New("shared memory is unavailable: /mnt/shared is not mounted")
	}
	root, err := git(repo, "rev-parse", "--show-toplevel")
	if err != nil {
		return location{}, errors.New("cannot find Git root")
	}
	remote, _ := git(repo, "remote", "get-url", "origin")
	name, err := identity(remote, root)
	if err != nil {
		return location{}, err
	}
	key := sha256.Sum256([]byte(name))
	return location{name, strings.HasPrefix(name, "path:"), filepath.Join(shared, "repo-memory", hex.EncodeToString(key[:]))}, nil
}

func timestamp() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func uuid() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, data)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func atomicWrite(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".write-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func exists(path string) bool { _, err := os.Stat(path); return err == nil }

func run(args []string, shared string, out io.Writer) error {
	global := flag.NewFlagSet("repo", flag.ContinueOnError)
	repo := global.String("repo", ".", "recipient checkout path")
	if err := global.Parse(args); err != nil {
		return err
	}
	args = global.Args()
	command := "location"
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}
	loc, err := locate(*repo, shared)
	if err != nil {
		return err
	}
	emit := func(value any) error { return json.NewEncoder(out).Encode(value) }
	inbox := filepath.Join(shared, "handoffs", filepath.Base(loc.Directory))
	if command == "location" {
		return emit(loc)
	}
	if command == "list" {
		entries, err := os.ReadDir(inbox)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		messages := []message{}
		for _, entry := range entries {
			if !entry.IsDir() || !validID.MatchString(entry.Name()) {
				continue
			}
			task := filepath.Join(inbox, entry.Name())
			var msg message
			if err := readJSON(filepath.Join(task, "metadata.json"), &msg); err != nil {
				return err
			}
			msg.Directory, msg.Status = task, "pending"
			if exists(filepath.Join(task, "claim")) {
				msg.Status = "claimed"
			}
			var result outcome
			if err := readJSON(filepath.Join(task, "outcome.json"), &result); err == nil {
				msg.Status = result.Status
			} else if !os.IsNotExist(err) {
				return err
			}
			messages = append(messages, msg)
		}
		return emit(map[string]any{"repo": loc.Repo, "path_fallback": loc.PathFallback, "messages": messages})
	}
	if command != "send" && command != "claim" && command != "finish" {
		return errors.New("unknown command")
	}
	var id string
	if command != "send" {
		if len(args) == 0 || !validID.MatchString(args[0]) {
			return errors.New("valid handoff ID required")
		}
		id, args = args[0], args[1:]
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	file := flags.String("file", "", "Markdown draft or report")
	token := flags.String("token", "", "claim token")
	status := flags.String("status", "", "done or blocked")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if len(flags.Args()) != 0 {
		return errors.New("unexpected arguments")
	}
	var content []byte
	if command == "send" || command == "finish" {
		content, err = os.ReadFile(*file)
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(content)) == "" {
			return errors.New("draft or report must not be empty")
		}
	}
	if command == "send" {
		if err := os.MkdirAll(inbox, 0700); err != nil {
			return err
		}
		id, err = uuid()
		if err != nil {
			return err
		}
		draft, err := os.MkdirTemp(inbox, ".draft-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(draft)
		msg := message{ID: id, Repo: loc.Repo, Created: timestamp()}
		if err := atomicWrite(filepath.Join(draft, "message.md"), content); err != nil {
			return err
		}
		if err := writeJSON(filepath.Join(draft, "metadata.json"), msg); err != nil {
			return err
		}
		msg.Directory = filepath.Join(inbox, id)
		if err := os.Rename(draft, msg.Directory); err != nil {
			return err
		}
		return emit(msg)
	}
	task := filepath.Join(inbox, id)
	if !exists(filepath.Join(task, "metadata.json")) {
		return errors.New("handoff not found in this repo's mailbox")
	}
	if command == "claim" {
		claim := filepath.Join(task, "claim")
		if err := os.Mkdir(claim, 0700); err != nil {
			return fmt.Errorf("cannot claim handoff; it may already be claimed: %w", err)
		}
		token, err := uuid()
		if err != nil {
			return err
		}
		if err := writeJSON(filepath.Join(claim, "owner.json"), ownership{token, timestamp()}); err != nil {
			return err
		}
		return emit(map[string]string{"id": id, "token": token, "message": filepath.Join(task, "message.md")})
	}
	var owner ownership
	if err := readJSON(filepath.Join(task, "claim", "owner.json"), &owner); err != nil {
		return err
	}
	if *token == "" || owner.Token != *token {
		return errors.New("claim token does not match")
	}
	if exists(filepath.Join(task, "outcome.json")) {
		return errors.New("handoff already has an outcome")
	}
	if *status != "done" && *status != "blocked" {
		return errors.New("status must be done or blocked")
	}
	if err := atomicWrite(filepath.Join(task, "result.md"), content); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(task, "outcome.json"), outcome{*status, timestamp()}); err != nil {
		return err
	}
	return emit(map[string]string{"id": id, "status": *status, "result": filepath.Join(task, "result.md")})
}

func main() {
	if err := run(os.Args[1:], "/mnt/shared", os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
