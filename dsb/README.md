# dsb

`dsb` is a host command. It needs Docker Sandboxes (`sbx`) on `PATH`.
From `dsb/`, run `go install ./cmd/dsb` with Go 1.25 or later.
Make sure Go’s install directory is on `PATH`. The dotfiles setup sets `GOBIN="$HOME/bin"` for this install.
After replacing the old `.zshrc`, start a fresh shell. `type dsb` should show the executable, not an alias.

## Source layout

- `cmd/dsb` reads settings and chooses what to mount and link for each sandbox.
- `internal/cli` routes commands and prints help. It calls the actions supplied by `cmd/dsb`.
- `internal/sandbox` wraps `sbx`. It takes mount and symbolic link data without knowing what
  the paths are for. It prepares host directories and writes mountable startup scripts.

The startup plan has two parts: a Bash renderer and a list of symbolic links.
`internal/sandbox` writes the Bash first, then appends the link commands.
The shell hook runs both parts inside the sandbox, where the source mounts exist.
On reopen, `cmd/dsb` passes the plan with the sandbox name and network rules in `sandbox.RunSpec`.

From `dsb/`, run `go test ./...` to test all packages.
The launcher tests use a fake `sbx`. They do not test live Docker mounts or provider access.
Use `go run ./cmd/dsb --help` to try the CLI without installing it.

## Host tasks

Run `dsb task new 01-fix-cache` from any folder in a Git working tree with a commit.
This creates `~/.dsb/shared/tasks/01-fix-cache/` and opens `TASK.md` in `$EDITOR` (default: `nvim`).
Write the prompt in that file. Keep inputs and outputs beside it.
To record extra read-only directories, pass their paths after the task name.
`DOCKER_SBX_SHARED_DIR` can override the shared directory.

Edit `settings.json` to change the task. The workspace defaults to the Git root.
The base defaults to the current commit. To track a local branch instead, replace the SHA with a branch name such as `main`.
The model and thinking level come from `shared/.pi/agent/settings.json` in the configured dotfiles directory.
A model-specific thinking level takes precedence over the default.

Set `ready` to `true` when the task is ready. Run `dsb task list` to see tasks in folder-name order.
Rename a folder to change its order. This does not change the stored branch name.
Run `dsb task open` to pick a folder by number, or `dsb task open 01-fix-cache` to open it directly.
### Run ready tasks

Run `dsb task run` to pick a task by number, or `dsb task run NAME` to run one ready task directly. Run `dsb auto` to process the ready tasks listed at startup in folder order.
Both commands run in the foreground. Neither prompts for input.
The default deadline is 30 minutes per task. Set `"timeout": "1h"` in `settings.json` to change it.

Each run creates a named clone from the project template. The template must include Pi and configured Git author details.
Pi runs in print mode on the requested branch. It can read `/mnt/task` and the numbered read-only mounts at `/mnt/input/1`, `/mnt/input/2`, and so on.
It can write artifacts to `/mnt/output`, which maps to `output/` in the task folder.
Task runs do not mount the interactive dotfiles at `/mnt/dotfiles` or link Pi settings and skills from that directory.
On success, dsb creates the result branch in the host repository without checking it out or pushing it. It then removes the sandbox.
On failure, dsb attempts to stop the sandbox and continues with the next task. Ctrl-C stops the queue.

Read `run/result`, `run/stderr`, and `run/response.md` in the task folder to review a run.
`run/sandbox` contains its sandbox name. A `run/` directory prevents automatic retries, even after interruption.
Before retrying, recover any changes from a retained sandbox and move the `run/` directory aside. Choose a new branch if the old one already exists.

## Interactive use

1. Run `dsb base` to create or reopen the build sandbox.
2. Allow install traffic for that sandbox only. Run `docker/main.sh` inside it. Do not log in there.
3. Remove the temporary network rule on the host. Check the effective policy before saving.
4. Run `dsb base save`.
5. Run `dsb` in a project directory. Sign in to Pi there. The project gets the shared login directory, not the build sandbox.

The Go launcher keeps the existing sandbox names. Interactive projects use the directory basename, so use distinct basenames.
Use `dsb open [MOUNT ...]` to pass extra mounts when you first create a project sandbox.
When `dsb` creates a sandbox, it writes a startup script under `~/.dsb/runtime`.
It mounts that directory read-only in the sandbox. Keep it while the sandbox exists.
The shell hook loads the script from a fixed path. It does not load scripts from the current project.
The script links Pi to the shared auth file if no local auth file exists.
If the mount is missing or a local file blocks the link, see [shared Pi login recovery](../README.md#shared-pi-login).

Startup rules are fixed when a sandbox is created. Recreate the sandbox to apply changes to its startup rules or mounts.

## Remove sandboxes

Run `dsb nuke` to stop and remove all local sandboxes, including the base and sandboxes not created by `dsb`.
`sbx` asks for confirmation. It may refuse to remove sandboxes that are in use.
This does not remove templates, host mount directories, or retained persistent volumes.

## Host settings

`dsb` uses the defaults below. Set a nonempty environment variable to override a default.

| Setting | Default |
| --- | --- |
| `DOCKER_SBX_BASE_NAME` | `tom-base` |
| `DOCKER_SBX_TEMPLATE` | `tom-dev:v1` |
| `DOCKER_SBX_DOTFILES_DIR` | `~/personal/code/dotfiles-setup` |
| `DOCKER_SBX_SHARED_DIR` | `~/.dsb/shared` |
| `DOCKER_SBX_PI_AUTH_DIR` | `~/.dsb/.pi/auth` |
| `DOCKER_SBX_ALLOWED_DOMAINS` | `auth.openai.com,proxy.golang.org,sum.golang.org,storage.googleapis.com,registry.npmjs.org` |

Every `dsb` project or base launch applies a network allowlist before it starts.
Use a comma-separated list of domain names to replace the default. For example:

```sh
export DOCKER_SBX_ALLOWED_DOMAINS=auth.openai.com,api.openai.com
```
The launcher adds rules. It does not remove rules from earlier launches when you change the list.
Check that your host policy permits only the provider endpoints you need.

Existing sandboxes keep their original startup scripts and mounts. Recreate them to remove the old Go cache mount and offline Go settings.
Go may need other hosts for private modules or direct dependencies. npm packages may also download files from other hosts.
The Go module proxy may redirect downloads to `storage.googleapis.com`. Allowing that host also permits access to other Google Cloud Storage buckets.
