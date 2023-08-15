import { FormEvent, useContext } from 'react';
import WebSocketContext from '../WebSocketContext';
import { Role, State } from '../enums';

interface PromptFormProps {
    gameState: State,
    playerRole: Role,
}

function PromptForm(props: PromptFormProps) {
    const ws = useContext(WebSocketContext);
    const formActive = props.gameState === State.GettingPrompt && props.playerRole === Role.Muse;
    const className = formActive ? "" : "inactive";

    const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (ws !== null) {

            ws.send(JSON.stringify({
                type: "prompt",
                data: { prompt: e.currentTarget.prompt.value },
            }));
            //TODO disable form?
        } else {
            console.error("cannot send prompt: no WebSocket")
        }
    };
    return (
      <div id="prompt-form-component" className={className}>
        <form id="prompt-form" onSubmit={handleSubmit}>
            <fieldset disabled={!formActive}>
                <input type="text" name="prompt" placeholder="Enter prompt here"></input>
                <input type="submit" value="Submit"></input>
            </fieldset>
        </form>
      </div>
    );
}

export { PromptForm }
