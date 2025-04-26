### GoToSocial kalaclista modified edition

The fork of [gotosocial](https://github.com/superseriousbusiness/gotosocial) for my personal instance.

#### Features

- Improve federation compatibility for doesn't support `AUTHORIZED_FETCH` instance
  - this feature make more compatibility to federate other instances
  - however, this featur make weakness about GoToSocial privacy guards
- Keep remmote emojis forever
  - this feature keeps remote emojis infinity times
  - this feature patch to vanish of remote emojis, but it has risk about copyrights
- For personal hack for broken databases
  - my instance has broken tables about posted by client
  - this is workaround for cannot looking the old posts

#### Configurations

- `kalaclista-allowed-unauthorized-get` (default: false)
- `kalaclista-keep-emojis-forever` (default: false)

#### Development

```shell
# track upstream
$ git pull upstream main
$ git push origin main

# make upstream branch
$ git switch -c gotosocial-vX.Y.Z {commit hash}

# make topic branches
$ git switch -c kalaclista-{topic}-vX.Y.Z gotosocial-vX.Y.Z
$ nvim .

# merge to kalaclista-vX.Y.Z
$ git switch -c kalaclista-vX.Y.Z gotosocial-vX.Y.Z
$ git rebase kalaclista-{topic}-vX.Y.Z # do repeat per branches

# finally, push to publich
$ git push origin kalaclista-vX.Y.Z
```

#### Maintainer

OKAMURA Naoki aka nyarla / kalaclista - [@nyarla@kalaclista.com](https://kalaclista.com/@nyarla)
