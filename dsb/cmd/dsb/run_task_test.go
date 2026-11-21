package main

import (
	"io"
	"reflect"
	"testing"
)

func TestPublishTaskBranch(t *testing.T) {
	branch := taskBranch{ref: "refs/heads/task/demo", remote: "sandbox-demo", base: "base"}
	for _, tc := range []struct {
		name   string
		status string
		ref    string
		fail   bool
	}{
		{name: "clean task branch", ref: branch.ref},
		{name: "dirty tree", status: "?? notes", ref: branch.ref, fail: true},
		{name: "other branch", ref: "refs/heads/main", fail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inside := func(out io.Writer, args ...string) error {
				var result string
				switch args[1] {
				case "status":
					result = tc.status
				case "symbolic-ref":
					result = tc.ref + "\n"
				case "rev-parse":
					result = "tip\n"
				default:
					t.Fatalf("unexpected sandbox command: %v", args)
				}
				_, err := io.WriteString(out, result)
				return err
			}

			var calls [][]string
			git := func(args ...string) (string, error) {
				calls = append(calls, args)
				return "", nil
			}
			err := publishTaskBranch(branch, inside, git)
			if tc.fail {
				if err == nil || len(calls) != 0 {
					t.Fatalf("expected rejection before host git; error %v, calls %v", err, calls)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := [][]string{
				{"fetch", "--no-tags", branch.remote, branch.ref},
				{"merge-base", "--is-ancestor", branch.base, "tip"},
				{"update-ref", branch.ref, "tip", ""},
			}
			if !reflect.DeepEqual(calls, want) {
				t.Errorf("host git calls: %v; want %v", calls, want)
			}
		})
	}
}
