import logo from './logo.svg';
import './App.css';
import { useEffect, useState } from 'react';

function App() {
  const [benchmarks, setBenchmarks] = useState([])
  const [websocket, setWebsocket] = useState(new WebSocket("ws://localhost:8000"))

  const testWS = () => {
    websocket.send("message;hello\n")
    websocket.send("benchmark;http://go.dev")
  }

  useEffect(() => {
    console.log('use effect')
    websocket.addEventListener('open', event => {
      console.log('WEBSOCKET OPEN', event)
    })

    websocket.addEventListener('message', event => {
      console.log('MESSAGE RECEIVED', event)
      setBenchmarks([...benchmarks, event.data])
    })

    websocket.addEventListener('error', event => {
      console.log('WEBSOCKET ERROR', event)
    })
  }, [websocket, benchmarks])

  useEffect(() => {
    websocket.addEventListener('close', event => {
      console.log('WEBSOCKET CLOSED, RECONNECTING')
      setTimeout(() => {
        setWebsocket(new WebSocket("ws://localhost:8000"))  
      }, 1000);
    })
  }, [websocket])
  
  return (
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.js</code> and save to reload.
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
        <button onClick={testWS}>Press me</button>
        {benchmarks.map((benchmark, i) => {
          return <p key={i}>{benchmark}</p>
        })}
      </header>
    </div>
  );
}

export default App;
