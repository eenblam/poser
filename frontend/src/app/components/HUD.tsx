import '../enums';
import { Role, State } from '../enums';
//import './HUD.css';

interface HudProps {
  gameState: State,
  currentPlayer: number,
  playerRole: Role,
  prompt: string,
}

function HUD(props: HudProps) {
  let roleHTML = props.gameState === State.Waiting ? "" : (<div id="">Your role: {props.playerRole}</div>);
  let gameHTML = props.gameState === State.Waiting  || props.gameState === State.GettingPrompt ? "" :
    (<div id="gameWrapper">
      <div id="">Current player: #{props.currentPlayer}</div>
      <div id="">Prompt: {props.prompt}</div>
    </div>);
  return (
  <div id="hud-div">
    <div id="">State: {props.gameState}</div>
    {roleHTML}
    {gameHTML}
  </div>
  );

}

export { HUD }
