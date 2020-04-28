[![Build Status](https://travis-ci.com/Yuruh/encrypted-diary.svg?branch=master)](https://travis-ci.com/Yuruh/encrypted-diary)
[![Version](https://img.shields.io/github/v/tag/yuruh/encrypted-diary)](https://github.com/Yuruh/encrypted-diary/releases)
[![codecov](https://codecov.io/gh/Yuruh/encrypted-diary/branch/master/graph/badge.svg)](https://codecov.io/gh/Yuruh/encrypted-diary)


# Goal

Building a diary solution where every entry is encrypted using the user's password.
It should be hosted on my servers, where multiple users could register, and offer the ability to self host, using a docker image and / or clear instructions

# Encryption process

To create encryption key : PBKDF2 hashing of user password

To encrypt / decrypt journal entries: AES-256

**All encryption must be done client side**

Only 2 routes should require password: login, and change-password (which has to rewrite all user journal entries)

# Road map

## Security

- Temporary login. Token no longer than 2 hour and set by client, and client should auto destroy encryption key when session expires.
- 2FA https://github.com/dgryski/dgoogauth


- (later on) External auth server that implements OAuth2 (https://pragmaticwebsecurity.com/files/cheatsheets/oauth2securityfordevelopers.pdf)
https://github.com/ory/hydra

## App Features

- Medias for each entry. Images for start. Should also be client-side encrypted. Use CDN (https://www.cloudflare.com/fr-fr/plans/, seems free)
        --> Done with labels avatars, but i'm not sure i can use a CDN with client side  encryption 

- Entry search (will be done when i've written enough entries)
# Resources

https://core.telegram.org/techfaq#q-how-does-end-to-end-encryption-work-in-mtproto
