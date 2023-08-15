import { FormEvent, useContext } from 'react';
import WebSocketContext from '../WebSocketContext';
import { State } from '../enums';

interface StartFormProps {
    gameState: State,
    playerNumber: number,
}

function StartForm(props: StartFormProps) {
    const ws = useContext(WebSocketContext);
    const formActive = props.gameState === State.Waiting && props.playerNumber === 1;
    const className = formActive ? "" : "inactive";

    const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (ws !== null) {

            ws.send(JSON.stringify({
                type: "start"
            }));
            //TODO disable form?
        } else {
            console.error("cannot send start: no WebSocket")
        }
    };
    return (
      <div id="start-form-component" className={className}>
        <form id="start-form" onSubmit={handleSubmit}>
            <fieldset disabled={!formActive}>
                <input type="submit" value="Start"></input>
            </fieldset>
        </form>
      </div>
    );
}

export { StartForm }
