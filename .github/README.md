## Kalaclista-flavoured GoToSocial

This is the fork of [GoToSocial](https://codeberg.org/superseriousbusiness/gotosocial) for my personal instance.

This fork is made for my own instance, I not recommended to use other users.

### Flavours

#### `kalaclista-turnoff-authorized-fetch` (defaut: true)

- this feature make to more compatibility for between other instance federation
- this feature make weakness to privacy guard of original GoToSocial

#### The quirk fix to my broken database

- my instance lost to the client application information from the old posts
- this is happened by my miss-operation to database, it cannot fix it by my skils

### Development notes

```shell
# sync upstream
$ git pull upstream main
$ git push origin main

# port my features
$ git co -b gotosocial-vA.B.C refs/tags/vA.B.C
$ git switch -c kalaclista-vA.B.C gotosocial-vA.B.C
$ nvim .

# push to my repo
$ git push origin kalaclista-vA.B.C
```

### Modifier

OKAMURA Naoki aka nyarla / kalaclista - [@nyarla@kalaclista.com](https://kalaclista.com/@nyarla)
