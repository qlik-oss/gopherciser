# VSCode snippets for gopherciser development in VSCode

In this folder there is a file called `gopherciser.code-snippets` containing snippets which can be used with VSCode.

### Installing the the snippets

Snippets can be "installed" 2 diffrent ways

1. On *nix systems it's recommended to create a symbolic link instead of copying the file to have it automatically be kept up to date. In the main repo folder create the symbolic link with the command `ln -s ../docs/vscode/gopherciser.code-snippets .vscode/gopherciser.code-snippets`.
2. Copy the file `gopherciser.code-snippets` from the `./docs/vscode` folder into the folder `.vscode` in the repo. This works on any OS, but snippet will not automatically be kept up to date and would need to be re-copied when there's an update.


### Using the template

Start writing the name of the snippet and press enter.

#### Action snippet

`action` snippet is a helper to automatically get a skeleton when creating new actions. This should be used the following way:

1. Create an empty file in package folder for the action, e.g. `scenario` folder, with the name of the new action, e.g. `dummy.go` for the action `dummy`.
2. Start writing `action` and press enter.
3. Change the action struct name if necessary.
4. Press *tab* to change package name (if needed)
5. Press *tab*  to change description
6. Press *tab*  to modify parameters