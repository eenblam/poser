import './PlayerList.css';

interface PlayerListProps {
  players: Player[];
}

class Player {
  constructor(
    public id: string,
    public playerNumber: number,
    public name: string,
    public owner: boolean,
  ) {}
}

function PlayerList(props: PlayerListProps) {
  const players = props.players;
  const listItems = players.map((player: Player) => {
    const className = `player-${player.playerNumber}`
    const userName = `Player #${player.playerNumber}`;
    return (<li key={player.playerNumber} className={className}>{userName}</li>);
  });
  return (
    <div id="playerlist-widget">
      <h2>Players</h2>
      <ul>{listItems}</ul>
    </div>
  );
}

export { Player, PlayerList }
