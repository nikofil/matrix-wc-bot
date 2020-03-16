# Matrix wordcount bot

A simple bot to check out golang's support for the Matrix client protocol and integration with libolm.

The bot will listen for events in the rooms that the user has joined and will reply to messages in the room of the form:

`!wc string`

by posting how many times that string has appeared in the room.

Can currently read events in encrypted rooms but can't reply. :)

To run:
`go build . && ./matrix-wc-bot http://localhost:8008 user pass matrix-wc-bot`
