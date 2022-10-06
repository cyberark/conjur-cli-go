# conjur-cli-go dev

`./start` creates a dev environment, including a CLI container that is already
logged into Conjur as the admin user.

`./stop` tears down the environment.

`./exec` connects to a CLI container if running, otherwise it runs `start`.