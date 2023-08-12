import { useRef, useState, useEffect } from 'react'
import { Canvas, DrawCallback, DrawCallbackContext } from './components/Canvas'
import { Chat, Message } from './components/Chat'
import { User, UserList} from './components/UserList'
import { Notification, Notifications } from './components/Notifications'
import WebSocketContext from './WebSocketContext'
import './App.css'

const wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = `${wsProtocol}//${location.host}${location.pathname.replace('/room/', '/ws/')}`;

const conn = new WebSocket(wsUrl, 'json');

function App() {
  let [_, setUserId] = useState<string>('...loading...');
  let [userList, setUserList] = useState<User[]>([]);
  let [messages, setMessages] = useState<Message[]>([]);
  let [notifications, setNotifications] = useState<Notification[]>([]);
  let [playerNumber, setPlayerNumber] = useState<number>(0);

  let connRef = useRef<WebSocket | null>(conn);
  let drawRef = useRef<DrawCallback>(new DrawCallback((_) => {
    console.error("draw callback called before initialization");
  }));

  useEffect(() => {
    // Translate http(s)://HOST/room/ID/ to ws(s)://HOST/ws/ID/
    /*
    let wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    let wsUrl = `${wsProtocol}//${location.host}${location.pathname.replace('/room/', '/ws/')}`;

    let conn = new WebSocket(wsUrl, 'json');
    */
    conn.onopen = (e) => {
      console.log(`wsConnection open to 127.0.0.1:8080`, e);
    };
    conn.onerror = (e) => {
      console.error(`wsConnection error `, e);
    };
    conn.onmessage = (e) => {
      let data = JSON.parse(e.data);
      const d = data.data; // may be undefined
      switch (data.type) {
        case 'connection':
          setUserId(d.id);
          setPlayerNumber(d.playerNumber);
          break;
        case 'ids':
          console.log(`Ids: ${data.ids}`);
          // Note: playerNumber (idx) is 1-indexed, not 0
          let users = data.ids.map((id: string, idx: number) => new User(id, idx+1, "", false))
                              .filter((u: User) => u.id !== ""); // ignore empty slots
          setUserList(users);
          break;
        case 'chat':
          let m = new Message(d.id, d.playerNumber, d.user, d.timestamp, d.text);
          console.log(`Chat: ${m.toWSMessage()}`);
          setMessages((messages) => [...messages, m]);
          break;
        case 'draw':
          drawRef.current.callback(data.data);
          break;
        case 'notification':
          let n = new Notification(d.timestamp, d.message, d.isError)
          // We can do .sort((a,b) => a.timestamp-b.timestamp) if we want timestamp ordering, but for now order of arrival seems best.
          setNotifications((ns) => [...ns, n]);
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
      <DrawCallbackContext.Provider value={drawRef.current}>
        <Notifications notifications={notifications}/>
        <UserList users={userList} />
        <Chat messages={messages} />
        <Canvas playerNumber={playerNumber}/>
      </DrawCallbackContext.Provider>
      </WebSocketContext.Provider>
    </>
  )
}

export default App
