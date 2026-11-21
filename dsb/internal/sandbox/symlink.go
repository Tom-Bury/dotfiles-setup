package sandbox

import (
	"io"
	"strings"
	"text/template"
)

// Symlink describes a link to create inside the sandbox when its source exists.
type Symlink struct {
	Source string // Absolute path inside the sandbox.
	Target string // Absolute path, or relative to the sandbox user's home.

	// PreserveExisting keeps any target, including a broken link.
	// Otherwise, the command replaces files and links but keeps real directories.
	PreserveExisting bool

	// RequireRegularSource skips sources that are not regular files.
	RequireRegularSource bool
	// SkipIfParentSymlink avoids writing through a symlink at the target's parent.
	SkipIfParentSymlink bool
}

var symlinkTemplate = template.Must(template.New("symlink").Funcs(template.FuncMap{
	"isAbsolute": func(s string) bool { return strings.HasPrefix(s, "/") },
	"quote": func(s string) string {
		return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
	},
}).Parse(`(
  source={{quote .Source}}
  target={{if not (isAbsolute .Target)}}"$HOME"/{{end}}{{quote .Target}}
  [[ {{if .RequireRegularSource}}-f{{else}}-e{{end}} "$source" ]] || exit 0
{{- if .SkipIfParentSymlink}}
  [[ -L "$(dirname "$target")" ]] && exit 0
{{- end}}
{{- if .PreserveExisting}}
  [[ -e "$target" || -L "$target" ]] && exit 0
  mkdir -p "$(dirname "$target")" && ln -sT "$source" "$target"
{{- else}}
  [[ -L "$target" && "$(readlink "$target")" == "$source" ]] && exit 0
  [[ -d "$target" && ! -L "$target" ]] && exit 0
  mkdir -p "$(dirname "$target")" && ln -sfnT "$source" "$target"
{{- end}}
)`))

// WriteShell writes one Bash command for use with Linux sandbox tools.
// The command returns success when there is no link to create.
func (link Symlink) WriteShell(w io.Writer) error {
	return symlinkTemplate.Execute(w, link)
}
