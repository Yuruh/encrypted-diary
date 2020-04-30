[![Build Status](https://travis-ci.com/Yuruh/encrypted-diary.svg?branch=master)](https://travis-ci.com/Yuruh/encrypted-diary)
[![Version](https://img.shields.io/github/v/tag/yuruh/encrypted-diary)](https://github.com/Yuruh/encrypted-diary/releases)
[![codecov](https://codecov.io/gh/Yuruh/encrypted-diary/branch/master/graph/badge.svg)](https://codecov.io/gh/Yuruh/encrypted-diary)


# Encrypted diary

A personal diary where every entry is encrypted using the user's password as encryption key.

## Self Host

You may self host this project.

TODO : --> explain dk compose, .env, ovh / postgresql

## About Encryption

When a user logs in, an encryption key is created using his password and [PBKDF2](https://en.wikipedia.org/wiki/PBKDF2). Your data is then encrypted / decrypted using [AES](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard).

This means that even if the database were to be compromised, your personal data would be safe as long as your password is. This also mean that if you forget your password your data is lost.

For practical reasons, some data that could be considered personal is not encrypted.

Here's the list of encrypted data:
* Entries Content
* Labels Avatar

And here's what isn't - *and why*:
* User Email Address - *For login*
* Entries Title - *For Entry search*
* Labels Names - *For Entry / Label search*
* Entries Date - *For Entry search*

## Features
 
* Short-lived session. Maximum 1h and auto log out on session end.
* Virtual Keyboard to enter password and prevent key logging.
* Two Factors Authentication with [Time-based One Time Password](https://en.wikipedia.org/wiki/One-time_password#Time-synchronized) (TOTP)
* Entry edition using **Markdown** format with live preview.
* Labels to categorize each entry, find entries by theme and act as a preview of an entry content

## Road map

*Disordered*

* Additional 2FA Methods
* 2FA Recovery codes
* Entries Media
* Entry search
* Read only user for demo purposes

## Contributing

TODO

## Resources

https://core.telegram.org/techfaq#q-how-does-end-to-end-encryption-work-in-mtproto
