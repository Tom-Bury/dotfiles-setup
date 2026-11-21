package main

import (
	_ "embed"
	"io"
	"path"
	"strings"
	"text/template"

	"dsb/internal/sandbox"
)

//go:embed startup.sh.tmpl
var startupSource string

var startupTemplate = template.Must(template.New("startup").Funcs(template.FuncMap{
	"quote": func(value string) string {
		return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
	},
}).Parse(startupSource))

type startupPaths struct {
	Dotfiles string
	PiAuth   string
}

func startupPlan(dir string, paths startupPaths) sandbox.Startup {
	const agentDir = ".pi/agent"
	piSource := path.Join(paths.Dotfiles, "shared", agentDir)
	startup := sandbox.Startup{Dir: dir}
	if paths.Dotfiles != "" {
		startup.Links = append(startup.Links, sandbox.Symlink{
			Source: path.Join(paths.Dotfiles, "shared/.agents/skills"),
			Target: ".agents/skills",
		})
		for _, name := range []string{"AGENTS.md", "APPEND_SYSTEM.md", "settings.json", "keybindings.json", "prompts"} {
			startup.Links = append(startup.Links, sandbox.Symlink{
				Source:              path.Join(piSource, name),
				Target:              path.Join(agentDir, name),
				SkipIfParentSymlink: true,
			})
		}
	}
	startup.Links = append(startup.Links, sandbox.Symlink{
		Source:               path.Join(paths.PiAuth, "auth.json"),
		Target:               path.Join(agentDir, "auth.json"),
		PreserveExisting:     true,
		RequireRegularSource: true,
		SkipIfParentSymlink:  true,
	})
	startup.Script = func(w io.Writer) error {
		return startupTemplate.Execute(w, struct {
			Dotfiles string
			Yazi     string
		}{paths.Dotfiles, path.Join(paths.Dotfiles, "shared/yazi")})
	}
	return startup
}
