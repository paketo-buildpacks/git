# Paketo Git Buildpack
The Paketo Git Buildpack is a Cloud Native Buildpack that retreives `git` metadata and preforms `git` operations.

## Behavior
This buildpack will only participate if there is a valid `.git` directory in the application source directory.

The buildpack will do the following:

- Sets the `REVISION` environment variable, which is the commitish of HEAD, to be available for the build processes of other buildpacks and in the final running image.
- Creates custom `git` credential managers if it is provided with credentials through a binding.

## Bindings
The buildpack optionally accepts the following bindings:

### Type: `git-credentials`
|Key                   | Value   | Description
|----------------------|---------|------------
|`credentials` | `<formated git credentials>` | The credentials file should have the following format to conform with the [`git` credential structure](https://git-scm.com/docs/git-credential#IOFMT).
|`context` (optional) | `<url>` |The context is an [optional pattern](https://git-scm.com/docs/gitcredentials#_credential_contexts) as defined by `git`. If a context is not provided then the credentials given in the binding will be the default credentials the `git` uses when authenticating. A given context can only be used once for any group of bindings, if a context is given by two separate bindings the build will fail.
