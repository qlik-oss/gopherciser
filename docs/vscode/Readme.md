# VSCode snippets for gopherciser development in VSCode

In this folder there is a file called `gopherciser.code-snippets` containing snippets which can be used with VSCode.

### Installing the the snippets

Snippets can be "installed" 2 diffrent ways

1. Copy the file `gopherciser.code-snippets` from the `docs/vscode` folder into the folder `.vscode` in the repo. This works on any OS.
2. On *nix systems it's recommended to create a symbolic link instead of copying the file to have it automatically be kept up to date. In the main repo folder create the symbolic link with the command `ln -s ../docs/vscode/gopherciser.code-snippets .vscode/gopherciser.code-snippets`.

### Using the template

Start writing the name of the snippet and press enter.

`action`: Adds skeleton of a scenario action. This should be used the following way:

1. Create an empty file in `scenario` folder with the name of the new action, e.g. `dummy.go` for the action `dummy`.
2. Start writing `action` and press enter.
3. Change the action struct name if necessary, then press *tab* to write a description.