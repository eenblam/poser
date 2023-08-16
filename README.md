# poser

Poser is a web version of a Fake Artist Goes to New York, by Oink Games.
It's still a work in progress, more to come!

Right now it's just a simple Go app on the backend, with TypeScript and React on the frontend.

## Running locally
You can do the following to get a test environment running:

```bash
git clone https://github.com/eenblam/poser
cd poser
make dependencies
make dev # build frontend, then run backend
```

The app will be available at http://localhost:8080.
However, Chrome doesn't allow insecure websockets (even on localhost,)
so you'll have a much better time with Firefox.

## State of implementation
Current roadmap:

* [x] Basic backend (creating/joining/leaving rooms, chat, sharing drawing data)
* [x] Simple shared canvas
* [x] Basic game implementation: assign user colors and implement turns
* [ ] Advanced game implementation: voting on the end result, guess for the fake artist
* [ ] Make the UI look nicer
* [ ] UX: custom user names and color picker
* [ ] Gallery: choose to save your final work, content mod tools, etc.

Since multiple connections need access to the same room,
the current implementation keeps them in an in-memory store
built on Go's concurrency primitives.
I might replace this with Redis later,
but for now the deployment plan is to instead ensure load balancers
route all users in a room to the same backend service.

When I get around to the gallery feature,
I'll build that out on a PostgreSQL backend.

## State of play
Players can currently draw freely during the lobby after joining.
Player #1 is the room owner, and can start the game by pressing Start.
This will clear the canvas and lock it.

A "Muse" is then selected: this player will pick the word to draw.
Once the Muse's prompt is submitted, all other players will be informed,
except for the Poser.
The Poser is instead left to guess what the prompt is by the end of drawing.

Currently, this all works! I just need to add voting at the end of the game
so players can guess who the Poser is.

