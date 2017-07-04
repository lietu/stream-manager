# Stream manager

An utility for managing various things around live streams.

Primarily built for use on Twitch, for my personal use, but can probably with little effort be extended to other uses.


## Requirements

 - MongoDB https://www.mongodb.com/download-center?jmp=nav#community
 - Golang https://golang.org/dl/
 - Twitch account https://www.twitch.tv/
 - `settings.yaml` (copy `settings.example.yaml`)
 - Twitch Developer Application


## Structure

 - `config`: Simply configuration parsing
 - `database`: Some minimal database connection logic abstraction
 - `html-overlay`: Home for the core library for browser source overlays. Check `example.html` and `README.md` for more.
 - `inventory`: Heavily WIP start for purchasing various "power-ups" for the stream
 - `lametric`: Integrations to http://lametric.com/
 - `manager`: Most of the core logic, pretty solid.
 - `storage`: Server for browser source overlay -files
 - `streamservice`: Basic abstraction for various stream services (Twitch, Mixer, ...)
 - `twitch`: Twitch specific integration
 - `utils`: General utilities for use everywhere else


## OAuth and Developer Application

You'll need an OAuth token, and a Twitch app registered.

The OAuth token should have the `chat_login` -scope. Probably will use the `channel_subscriptions` -scope in the future
too. You can probably use the one https://twitchapps.com/tmi/ generates, just remove the `oauth:` -prefix.

Create your Developer Application at https://www.twitch.tv/settings/connections and get it's Client ID.


## Installation and setup

Make sure you have a working Go environment and run:

```
go get github.com/lietu/stream-manager
# Windows: cd %GOPATH%\src\github.com\lietu\stream-manager
# Others: cd $GOPATH/src/github.com/lietu/stream-manager
```


## Usage

Make sure your `settings.yaml` is good.

On Windows:

```
run
```

On others:

```
go build stream-manager.go
./stream-manager
```

You'll likely want to allow traffic at least from localhost to your port 60606 (or whatever you changed that).

Create your notification frontend (e.g. using `html-overlay/example.html` as a base) and set it up for OBS/Xsplit.


## Related

https://github.com/lietu/stream-manager-unity-frontend


## License

MIT and/or new BSD

Pick which one works better for you.
