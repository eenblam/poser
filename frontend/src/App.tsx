import { Component } from 'react'
import UserList from './components/UserList'
import { Chat, Message } from './components/Chat'
import WebSocketContext from './WebSocketContext'
import './App.css'

interface AppState {
  userId: string;
  userList: string[];
  messages: Message[];
}

class App extends Component<{}, AppState> {
  conn: WebSocket;

  constructor(props: AppState) {
    super(props);

    this.state = {
      userId: '...loading...',
      userList: [],
      messages: [],
    }

    // Translate http(s)://HOST/room/ID/ to ws(s)://HOST/ws/ID/
    let wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    let wsUrl = `${wsProtocol}//${location.host}${location.pathname.replace('/room/', '/ws/')}`;

    this.conn = new WebSocket(wsUrl, 'json');
    this.conn.onopen = (e) => {
      console.log(`wsConnection open to 127.0.0.1:8080`, e);
    };
    this.conn.onerror = (e) => {
      console.error(`wsConnection error `, e);
    };
    this.conn.onmessage = (e) => {
      let data = JSON.parse(e.data);
      switch (data.type) {
        case 'connection':
          this.setState({...this.state, userId: data.id});
          break;
        case 'ids':
          console.log(`Ids: ${data.ids}`);
          this.setState({...this.state, userList: data.ids});
          break;
        case 'chat':
          let d = data.data
          let m = new Message(d.id, d.user, d.timestamp, d.text);
          console.log(`Chat: ${m.toWSMessage()}`);
          this.setState({...this.state, messages: [...this.state.messages, m]});
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
  }


  render() {
  return (
    <>
      <h1>Poser</h1>
      <WebSocketContext.Provider value={this.conn}>
        <UserList users={this.state.userList} />
        <Chat messages={this.state.messages} />
      </WebSocketContext.Provider>
    </>
  )
  }
}

export default App
