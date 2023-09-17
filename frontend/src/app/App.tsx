import { useRef, useState, useEffect } from 'react'
import { Canvas, DrawCallback, DrawCallbackContext } from './components/Canvas'
import { Chat, Message } from './components/Chat'
import { Player, PlayerList} from './components/PlayerList'
import { Notification, Notifications } from './components/Notifications'
import { StartForm } from './components/StartForm'
import { PromptForm } from './components/PromptForm'
import { State, Role } from './enums';
import WebSocketContext from './WebSocketContext'
import './App.css'

// Translate http(s)://HOST/room/ID/ to ws(s)://HOST/ws/ID/
const wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = `${wsProtocol}//${location.host}${location.pathname.replace('/room/', '/ws/')}`;

const conn = new WebSocket(wsUrl, "json");

function App() {
  let [_, setUserId] = useState<string>('...loading...');
  let [playerList, setPlayerList] = useState<Player[]>([]);
  let [messages, setMessages] = useState<Message[]>([]);
  let [notifications, setNotifications] = useState<Notification[]>([]);
  let [playerNumber, setPlayerNumber] = useState<number>(0);
  let [gameState, setGameState] = useState<State>(State.Waiting);
  let [playerRole, setPlayerRole] = useState<Role>(Role.Artist);
  let [currentPlayer, setCurrentPlayer] = useState<number>(0);

  let connRef = useRef<WebSocket | null>(conn);
  let drawRef = useRef<DrawCallback>(new DrawCallback((_) => {
    console.error("draw callback called before initialization");
  }));

  useEffect(() => {
    conn.onopen = (e) => {
      console.log(`wsConnection open to ${wsUrl}`, e);
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
        case 'players':
          console.log(`Ids: ${d.ids}`);
          // Note: playerNumber (idx) is 1-indexed, not 0
          let users = d.ids.map((id: string, idx: number) => new Player(id, idx+1, "", false))
                              .filter((u: Player) => u.id !== ""); // ignore empty slots
          setPlayerList(users);
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
        case 'role':
          setPlayerRole(d.role);
          break;
        case 'state':
          console.log(`State: ${d.state}`);
          setGameState(d.state);
          break;
        case 'turn':
          console.log(`Current player: ${d.playerNumber}`);
          setCurrentPlayer(d.playerNumber);
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
      <div id="app-wrapper">
        <WebSocketContext.Provider value={connRef.current}>
        <DrawCallbackContext.Provider value={drawRef.current}>
          <div id="ui-wrapper">
            <Notifications notifications={notifications}/>
            <StartForm gameState={gameState} playerNumber={playerNumber} />
            <PromptForm gameState={gameState} playerRole={playerRole} />
            <PlayerList players={playerList} />
          </div>
          <Chat messages={messages} />
          <Canvas gameState={gameState} playerNumber={playerNumber} currentPlayer={currentPlayer}/>
        </DrawCallbackContext.Provider>
        </WebSocketContext.Provider>
      </div>
    </>
  )
}

export default App
