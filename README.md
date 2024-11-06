# dotfiles-setup

Repository to to kickstart a Mac or Linux machine, according to personal preferences.
Some of the setup is optional, depending on the values provided in the prompt.
These options are kept in a `.env` such that they don't need to be asked again for reruns.

## Script structure and usage

The setup scripts are split in a common and os specific part:

### Shared

- ZSH setup, including aliases, custom functions and Oh My ZSH + Starship
- Git configs and aliases
- [optional] Go setup, currently just installing GVM (Go Version Manager)
- [optional] NodeJS setup, currently just installing NVM (NodeJS Version Manager)

### OSX specific

- Run `osx/main.sh` which installs the shared bits, as well as:
  - MacOS system settings changes
  - Installing Homebrew
  - Installing apps and VSCode extensions defined in `osx/homebrew/Brewfile`
  - Syncs VSCode settings and keybindings

#### Manual extras

- Manually set up iTerm2 to use the preferences stored in this repo at `iterm2/com.googlecode.iterm2.plist`:
  - Open iTerm2 preferences
  - Go to General > Preferences
  - Check "Load preferences from a custom folder or URL" and set the path to the repo folder
- In Git Fork GUI, manually add the custom commands from `/git-fork-custom-commands`
- In Raycast, manually import the settings from `raycast/settings-export.rayconfig`
- Call `setup_ssh` to add your SSH keys to your SSH agent using the keychain

### Linux specific

- Firstly, `linux/install-zsh.sh` needs to be run to update & upgrade, and install ZSH
- Afterwards `linux/main.sh` can be run which installs the shared bits, as well as:
  - A limited set of packages using `apt`
  - VSCode extensions defined in `osx/homebrew/Brewfile`
  
#### Manual extras

- Manually sync the VSCode settings and keybindings

## Extra info

The script can be run multiple times, it should be idempotent.
The script will create backups before overwriting the following files:

- `~/.gitconfig`
- `~/.zshrc`
- `~/.config/starship.toml`
- [OSX] VSCode `~/Library/Application Support/Code/User/settings.json`
- [OSX] VSCode `~/Library/Application Support/Code/User/keybindings.json`

Backups are created in `~/.mac-setup-backup`
