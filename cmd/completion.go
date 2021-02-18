package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate command line completion script",
	Long: fmt.Sprintf(`To load completions:

Bash:

$ source <(%[1]s completion bash)

# To load completions for each session, execute once:
Linux:
  $ %[1]s completion bash > /etc/bash_completion.d/%[1]s
MacOS:
  $ %[1]s_osx completion bash > /usr/local/etc/bash_completion.d/%[1]s_osx

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

# You will need to start a new shell for this setup to take effect.

Fish:

$ %[1]s completion fish | source

# To load completions for each session, execute once:
$ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

Powershell:

PS> %[1]s.exe completion powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> %[1]s completion powershell > %[1]s.ps1
# and source this file from your powershell profile.
`, RootCmd.Name()),
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		switch args[0] {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	},
}
