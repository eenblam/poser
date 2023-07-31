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
    //console.log(JSON.parse(e.data));
    console.log(e);
};
