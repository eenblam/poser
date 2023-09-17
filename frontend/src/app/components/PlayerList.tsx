import { FormEvent, useContext } from 'react';
import WebSocketContext from '../WebSocketContext';
import { State } from '../enums';
import './PlayerList.css';

interface PlayerListProps {
  gameState: State,
  players: Player[],
}

class Player {
  constructor(
    public id: string,
    public playerNumber: number,
    public name: string,
    public owner: boolean,
    public votes: number,
  ) {}
}

function PlayerList(props: PlayerListProps) {
  const players = props.players;
  const ws = useContext(WebSocketContext);

  const formActive = [
    State.Voting, State.PoserWon, State.PoserWonByTie, State.PoserLost, State.PoserGuessing
  ].includes(props.gameState);
  const inactiveClassName = formActive ? "" : "inactive";

  const listItems = players.map((player: Player) => {
    const className = `player-${player.playerNumber}`
    const userName = `Player #${player.playerNumber}`;
    // Set up vote sidebar
    const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (ws !== null) {
            let vote = JSON.stringify({
              type: "vote",
              data: {
                vote: player.playerNumber
              }
            });
            console.log(`Sending vote: ${vote}`);
            ws.send(vote);
            //TODO disable form?
        } else {
            console.error("cannot send vote: no WebSocket")
        }
    };
    const voteElts = <form className={inactiveClassName} onSubmit={handleSubmit}>
      <fieldset disabled={!formActive}>
                <input id="player-{player.PlayerNumber}-submit" type="submit" value="Vote"></input>
      </fieldset>
      </form>
    const voteCount = <span className={inactiveClassName}> {player.votes} </span>;
    return (<li key={player.playerNumber} className={className}>
        {voteElts}
        {voteCount}
        <span>{userName}</span>
      </li>);
  });
  return (
    <div id="playerlist-widget">
      <h2>Players</h2>
      <ul>{listItems}</ul>
    </div>
  );
}

export { Player, PlayerList }
