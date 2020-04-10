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

# Envisioned Features

- Search by approximate title (https://www.postgresql.org/docs/9.1/fuzzystrmatch.html, fuzzy / approximate string matching)
- Search by tag, with tag management (crud, name / color) and autocompletion in editors


# Resources

https://core.telegram.org/techfaq#q-how-does-end-to-end-encryption-work-in-mtproto
