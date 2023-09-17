enum State {
  Waiting = "Waiting",
  GettingPrompt = "GettingPrompt",
  Drawing = "Drawing",
  Voting = "Voting",
  PoserGuessing = "PoserGuessing",
  PoserWon = "PoserWon",
  PoserWonByTie = "PoserWonByTie",
  PoserLost = "PoserLost",
}

enum Role {
  Artist = "Artist",
  Muse = "Muse",
  Poser = "Poser",
}

export {
    Role,
    State,
}
