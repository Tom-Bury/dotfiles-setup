#!/usr/bin/env zsh

# Adapted from https://thevaluable.dev/fzf-git-integration/

function git_add() {
	local -r prompt_add="  Stage > "
	local -r prompt_reset="  Unstage > "

	local -r git_root_dir=$(git rev-parse --show-toplevel)
	local -r git_unstaged_files="git ls-files --modified --deleted --other --exclude-standard --deduplicate $git_root_dir"

	local git_staged_files='git status --short | grep "^[A-Z]" | awk "{print \$NF}"'

	local -r git_reset="git reset -- {+}"
	local -r enter_cmd="($git_unstaged_files | grep {} && git add {+}) || $git_reset"

	local -r preview_status_label="[ 📝 Status ]"
	local -r preview_status="git -c color.status=always status --short"
	local -r preview_diff_label="[ 🔄 Diff ]"
	local -r preview_diff="git diff --color=always {} | sed "1,4d""
	local -r preview_blame_label="[ 🔍 Blame ]"
	local -r header=$(cat <<-EOF
		| OPT-S: Switch Stage/Unstage
		| Preview: CTRL-T: Status | CTRL-F: Diff | CTRL-B: Blame
		| OPT-E: Open in editor
		| Commit: OPT-C: Commit | OPT-A: Amend
	EOF
	)
	local -r add_header=$(cat <<-EOF
		$header
		| TAB: Select files
		| ENTER: Stage selected
		| OPT-P: Add patch
	EOF
	)

	local -r reset_header=$(cat <<-EOF
		$header
		| TAB: Select files
		| ENTER: Unstage selected
		| OPT-D: Discard file
	EOF
	)

	local -r mode_reset="change-prompt($prompt_reset)+reload($git_staged_files)+change-header($reset_header)+unbind(alt-p)+rebind(alt-d)"
	local -r mode_add="change-prompt($prompt_add)+reload($git_unstaged_files)+change-header($add_header)+rebind(alt-p)+unbind(alt-d)"

	eval "$git_unstaged_files" | fzf \
	--color='prompt:#af5fff,header:#262626,gutter:-1,pointer:#af5fff,marker:#87ff00' \
	--marker="> " \
	--pointer="⏺ " \
	--multi \
	--reverse \
	--no-sort \
	--prompt="$prompt_add" \
	--preview-label="$preview_diff_label" \
	--preview="$preview_diff" \
	--header "$add_header" \
	--header-first \
	--bind='start:unbind(alt-d)' \
	--bind="ctrl-t:change-preview-label($preview_status_label)" \
	--bind="ctrl-t:+change-preview($preview_status)" \
	--bind="ctrl-f:change-preview-label($preview_diff_label)" \
	--bind="ctrl-f:+change-preview($preview_diff)" \
	--bind="ctrl-b:change-preview-label($preview_blame_label)" \
	--bind='ctrl-b:+change-preview(git blame --color-by-age {})' \
	--bind="alt-s:transform:[[ \$FZF_PROMPT =~ '$prompt_add' ]] && echo '$mode_reset' || echo '$mode_add'" \
	--bind="enter:execute($enter_cmd)" \
	--bind="enter:+reload([[ \$FZF_PROMPT =~ '$prompt_add' ]] && $git_unstaged_files || $git_staged_files)" \
	--bind="enter:+refresh-preview" \
	--bind='alt-p:execute(git add --patch {+})' \
	--bind="alt-p:+reload($git_unstaged_files)" \
	--bind="alt-d:execute($git_reset && git checkout {+})" \
	--bind="alt-d:+reload($git_staged_files)" \
	--bind='alt-c:execute(git commit)+abort' \
	--bind='alt-a:execute(git commit --amend)+abort' \
	--bind='alt-e:execute(code -r {+})' \
	--bind='f1:toggle-header' \
	--bind='f2:toggle-preview' \
	--bind='ctrl-y:preview-up' \
	--bind='ctrl-e:preview-down' \
	--bind='ctrl-u:preview-half-page-up' \
	--bind='ctrl-d:preview-half-page-down'
}

alias gita='git_add'