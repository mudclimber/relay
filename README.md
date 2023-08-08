# relay (mudclimber)

This is a mudclimber library that lets you relay any TCP traffic to the Hub.

## Interface

It takes TCP traffic and feeds it into a nice ANSI box with buffered output as well as prompt reading functionality.

<img width="754" alt="image" src="https://github.com/mudclimber/relay/assets/118575040/9317212b-00db-40dc-885f-cead62884757">

```mermaid
flowchart LR
H["Hub"]
R["Relay"]
E["Endpoint"]

H --> R --> E
E --> R --> H
```

## Intercepting output

You can intercept your TCP traffic when a player connects to the game through The Hub.  This is useful is you
want to take an existing MUD and handle your own auth behind the scenes, since The Hub already has players
authenticated through Twitch.

## Parsing output

You can take any output and parse it and replace it with whatever you want. Useful if you want to remove
colors, telnet negotiation stuff, or just want to do something weird like capitalize everything that comes through.

## Architecture

```mermaid
flowchart TD
TP["Terminal Processor"]
FP["Frame Processor"]
GS["Game I/O"]
AS["Agent I/O"]
TE["TCP Endpoint"]
HC["Hub Connection (the player)"]

FP -- "Converts agent frames to terminal operations" --> TP
GS --> TE --> GS
GS --> TP --> GS
AS --> FP --> AS
HC --> AS --> HC
TP -- "Draws ANSI on agents" --> AS
```

## About mudclimber

If you are wondering what the context of this is, here are some links:

* [mudclimber.dev](https://mudclimber.dev) <- go here if you're new!
* [Twitch](https://twitch.tv/mudclimber)
* [Twitter](https://twitter.com/mudclimber)
