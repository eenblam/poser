import { useRef, useState, useEffect } from 'react'
import Canvas from './components/Canvas'
import UserList from './components/UserList'
import { Chat, Message } from './components/Chat'
import WebSocketContext from './WebSocketContext'
import './App.css'

function App() {
  let [_, setUserId] = useState<string>('...loading...');
  let [userList, setUserList] = useState<string[]>([]); // how to set type string[]?
  let [messages, setMessages] = useState<Message[]>([]); // how to set type Message[]?
  let connRef = useRef<WebSocket | null>(null);

  useEffect(() => {
        // Translate http(s)://HOST/room/ID/ to ws(s)://HOST/ws/ID/
    let wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    let wsUrl = `${wsProtocol}//${location.host}${location.pathname.replace('/room/', '/ws/')}`;

    // let conn = new WebSocket(wsUrl, 'json');
    let conn = new WebSocket(wsUrl, 'json');
    conn.onopen = (e) => {
      console.log(`wsConnection open to 127.0.0.1:8080`, e);
    };
    conn.onerror = (e) => {
      console.error(`wsConnection error `, e);
    };
    conn.onmessage = (e) => {
      let data = JSON.parse(e.data);
      switch (data.type) {
        case 'connection':
          setUserId(data.id);
          break;
        case 'ids':
          console.log(`Ids: ${data.ids}`);
          setUserList(data.ids);
          break;
        case 'chat':
          let d = data.data
          let m = new Message(d.id, d.user, d.timestamp, d.text);
          console.log(`Chat: ${m.toWSMessage()}`);
          setMessages((messages) => [...messages, m]);
          break;
        case undefined:
          console.log("Undefined message type");
          console.log(e);
          break;
        default:
          console.log(`Unknown message type: ${data.type}`);
          console.log(e);
      }
    };
    connRef.current = conn;
    // Return a cleanup function to cleanly close the connection
    return () => {
      conn.close();
      connRef.current = null;
    };
  }, []);

  return (
    <>
      <h1>Poser</h1>
      <WebSocketContext.Provider value={connRef.current}>
        <UserList users={userList} />
        <Chat messages={messages} />
        <Canvas/>
      </WebSocketContext.Provider>
    </>
  )
}

export default App
