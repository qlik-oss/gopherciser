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
	Long: `To load completions:
  Bash:
    $ source <(gopherciser completion bash)

    # To load completions for each session, execute once:
    Linux:
      $ gopherciser completion bash > /etc/bash_completion.d/gopherciser
    MacOS:
      $ gopherciser_osx completion bash > /usr/local/etc/bash_completion.d/gopherciser_osx

  Zsh:
    # If shell completion is not already enabled in your environment you will
    # need to enable it. You can execute the following once:

    $ echo "autoload -U compinit; compinit" >> ~/.zshrc

    # To load completions for each session, execute once:
    $ gopherciser completion zsh > "${fpath[1]}/_gopherciser"

    # You might want to put the completion script (_gopherciser) at another
    # path within the fpath. To view the fpath, execute:
    $ echo $fpath

    # You will need to start a new shell for this setup to take effect.

  Fish:
    $ gopherciser completion fish | source

    # To load completions for each session, execute once:
    $ gopherciser completion fish > ~/.config/fish/completions/gopherciser.fish

  Powershell:
    PS> gopherciser.exe completion powershell | Out-String | Invoke-Expression

    # To load completions for every new session, run:
    PS> gopherciser completion powershell > gopherciser.ps1
    # and source this file from your powershell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
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
