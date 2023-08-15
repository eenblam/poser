import { ChangeEvent, FormEvent, useContext, useState } from 'react';
import WebSocketContext from '../WebSocketContext';

interface ChatProps {
    messages: Message[];
}
  
function Chat(props: ChatProps) {
    const messages = props.messages;
    const chatMessages = messages.sort((a,b) => a.timestamp-b.timestamp).map((m: Message) => {
        const className = `player-${m.playerNumber}`
        return (<p key={m.id} className={className}><ChatItem message={m} /></p>);
    }
    );

    return (
        <div id="chat-widget">
            <h2>Chat</h2>
            <div>{chatMessages}</div>
            <ChatInput />
        </div>
    );
}

interface ChatItemProps {
    message: Message;
}

function ChatItem(props: ChatItemProps) {
    const m = props.message;
    return (
        <div>
            <p>Player #{m.playerNumber}: {m.text}</p>
        </div>
    );
}

class Message {
    constructor(
        public id: string,
        public playerNumber: number,
        public user: string,
        public timestamp: number,
        public text: string,
    ) {}

    string() {
        return JSON.stringify(this);
    }

    toWSMessage() {
        return JSON.stringify({
            type: 'chat',
            data: this,
        })
    }
}

function ChatInput() {
    const ws = useContext(WebSocketContext);
    let [message, setMessage] = useState('');

    const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        if (ws !== null) {
            // Most of these are overwritten on the server side except timestamp and message
            let m = new Message('', 0, '', Date.now(), message)
            ws.send(m.toWSMessage())
            setMessage('');
        } else {
            console.error("cannot send chat: no WebSocket")
        }
    }

    const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
        setMessage(e.target.value);
    }

    return (
        <form
            onSubmit={handleSubmit}
            className="chat-input-form">
            <input
                onChange={handleChange}
                type="text"
                name="chat-message"
                value={message}
                placeholder="Type a message and hit Enter"
                />
            <button>Send</button>
        </form>
    );
}

export {
    Chat,
    ChatItem,
    Message,
}
