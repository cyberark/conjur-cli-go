# conjur-cli

## Upgrading from docker based CLI

See the table below for changes in version 8.x

| docker based<br>CLI command         | Replacement<br>Version 8 | Description |
|----------------------------------|     ---------------------|----------------------------------|
| authn  authenticate | [authenticate](#conjur-authenticate) | `conjur authn authenticate` is now `conjur authenticate`. [Some options have changed.](#conjur-authenticate)
| authn  login        | [login](#conjur-login)   | `conjur authn login` is now `conjur login`. Some options have changed.
| authn  logout       | [logout](#conjur-logout) | `conjur authn logout` is now `conjur logout`.
| authn  whoami       | [whoami](#conjur-whoami) | `conjur authn whoami` is now `conjur whoami`.
| check      | check |
| env        | not supported in 8.x   | [Can use shell scripts or Summon](#conjur-env)
| help       | help                   |
| host       | [host](#conjur-host)   | Options have changed
| hostfactory| [hostfactory](#conjur-hostfactory)| Added flags for token and id, changed the --duration flags, see [below](#conjur-hostfactory)
| init       | [init](#conjur-init)   | Options have changed.
| ldap-sync  | not supported in 8.x   | [Can use curl](#conjur-ldap-sync)
| list       | [list](#conjur-list)   | Removed the --raw-annotations and other minor changes.
| policy     | [policy](#conjur-policy)    | Added replace and append as separate subcommands.
| pubkeys    | [pubkeys](#conjur-pubkeys)  | the optional show command is not needed or supported.
| resource   | [resource](#conjur-resource)| permitted_roles is changed to permitted-roles
| role       | [role](#conjur-role)        | Minor changes to options
| show       | not supported in 8.x        | Replaced by show options in individual commands.
| user       | [user](#conjur-user)        | update_password and  rotate_api_key have changed to change-password and rotate-api-key. Small changed to option args.
| variable   | [variable](#conjur-variable)| Arguments have changed.

### `conjur authenticate`
```
The `conjur authenticate` command is simplified in 8.x by removing the `authn` keyword.

The `-f, --filename=filename` option in docker based CLI is no longer supported
The --no-header option is not available, the equivalent option is to omit the --header option.
```

### `conjur login`
```
The `conjur login` command is simplified in 8.x by removing the `authn` keyword.
The -u/--user option in docker based CLI is changed to -i/--id in 8.x

```
### `conjur logout`
```
The `conjur logout` command is simplified in 8.x by removing the `authn` keyword.

```

### `conjur whoami`
```
The `conjur whoami` command is simplified in 8.x by removing the `authn` keyword.
```

### `conjur env`
```
`conjur env` is not not supported in 8.x
As a replacement the `conjur variable` command can be used to set an
environmental variable.

URL=$(conjur variable value  test-secrets-provider-rotation-app-db/url) eval 'echo The url is $URL'
The url is postgresql://test-app-backend.app-test.svc.cluster.local:5432/test_app

```
Another option is to use [Summon](https://cyberark.github.io/summon/)
```
summon -p summon-conjur --yaml 'url: !var  test-secrets-provider-rotation-app-db/url' eval 'echo The url is $URL'
```

### `conjur host`
```
host layers is not supported in 8.x
host rotate_api_key is changed to host rotate-api-key
--host option is changed to --id

```

### `conjur hostfactory`
```

  For the token create command the duration is changed to a single duration flag.
  Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
  conjur hostfactory tokens create [flags]
  --duration-days, --duration-hours and --duration-minutes are being deprecated.

  The hostfactory is proceeded with the --hostfactory option
 -f, --hostFactory string   Host Factory id
  The -c option is short for --cide, as it was previously short for --count

```

### `conjur init`
```
-c, --certificate has changed to -c, --ca-cert
additional options have been added in 8.x

```

### `conjur ldap-sync`
```
The conjur ldap-sync is not supported in version 8.x. The curl command and the rest API can be used to
access the ldap-sync data. Use the following command with the specific cacert and conjur-fqdn to
get the ldap-sync data and save it to a file.
curl -v \
  --cacert "/root/conjur-<account>.pem" \
  --header "$(conjur authenticate -H)" \
  https://<conjur-fqdn>/api/ldap-sync/policy?config_name=default
  | jq '.policy' --raw-output \
  > ldap-sync.yaml
```

### `conjur list`
````
Minor changes to existing options.
The -r option is now short for --role
````

### `conjur policy`
```
The conjur policy command has added two additional subcommands replace and update
to replace the command options in the docker based CLI.
The branch and filename need to be preceeded with the -b and -f option flags.
Examples:
- conjur policy load -b staging -f /policy/staging.yml
```

### `conjur pubkeys`
```
The docker based CLi had an optional show command, this is not supported.
```

### `conjur resource`
```
The command permitted_roles is changed to permitted-roles
```

### `conjur role`
```
For the conjur role membership command, the --system option is not supported
```

### `conjur user`
```
The command update_password has changed to change-password The command rotate_api_key has changed to rotate-api-key.
For the rotate-api-key command the --user option is changed to --id
```

### `conjur variable`
```
The conjur variable value <VARIABLE> has changed to conjur variable get -i <VARIABLE>.
The conjur variable value <VARIABLE> <VALUE> has changed to conjur variable value
set -i <VARIABLE> -v <VALUE>
```