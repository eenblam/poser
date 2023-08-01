import { Component } from 'react'
import UserList from './components/UserList'
import './App.css'

interface AppState {
  userId: string;
  userList: string[];
}

class App extends Component<{}, AppState> {
  constructor(props: AppState) {
    super(props);

    this.state = {
      userId: '...loading...',
      userList: [],
    }

    // Translate http(s)://HOST/room/ID/ to ws(s)://HOST/ws/ID/
    let wsProtocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
    let wsUrl = `${wsProtocol}//${location.host}${location.pathname.replace('/room/', '/ws/')}`;

    const wsConnection = new WebSocket(wsUrl, 'json');
    wsConnection.onopen = (e) => {
      console.log(`wsConnection open to 127.0.0.1:8080`, e);
    };
    wsConnection.onerror = (e) => {
      console.error(`wsConnection error `, e);
    };
    wsConnection.onmessage = (e) => {
      let data = JSON.parse(e.data);
      switch (data.type) {
        case 'connection':
          this.setState({...this.state, userId: data.id});
          break;
        case 'ids':
          console.log(`Ids: ${data.ids}`);
          this.setState({...this.state, userList: data.ids});
          break;
        case undefined:
          console.log("Undefined message type");
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
      <UserList users={this.state.userList} />
    </>
  )
  }
}

export default App
