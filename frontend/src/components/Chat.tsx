import { ChangeEvent, FormEvent, useContext, useState } from 'react';
import WebSocketContext from '../WebSocketContext';

interface ChatProps {
    messages: Message[];
}
  
function Chat(props: ChatProps) {
    const messages = props.messages;
    const listItems = messages.sort((a,b) => a.timestamp-b.timestamp).map((m: Message) =>
        <li key={m.id}><ChatItem message={m} /></li>
    );

    return (
        <div>
            <h2>Chat</h2>
            <ul>{listItems}</ul>
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
            <p>{m.user}: {m.text}</p>
        </div>
    );
}

class Message {
    constructor(
        public id: string,
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
            let m = new Message('', '', Date.now(), message)
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
