# shh

simple cli for populating env vars with secrets

## Installation

```sh
brew tap robbiecloset/shh && brew install shh
```

## Usage

`shh` is dependent upon a file named `.env.secrets` in the directory from which you are running the command. A test / example file is included in the root of this project.

The format it expects is as follows:

```sh
ENV_VAR_NAME=asm/secret/name
```

The value on the left of the equal sign is the name of the variable injected into your env. The value on the right of the equal sign is the name of the secret in aws secretsmanager (the only secrets store supported at the moment).

Plain text secrets work, as do maps.
