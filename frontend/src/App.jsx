import "./App.css";
import { useEffect, useState } from "react";

function App() {
	const [url, setUrl] = useState("");
	const [benchmarks, setBenchmarks] = useState([]);
	const [websocket, setWebsocket] = useState(
		new WebSocket("ws://localhost:8000"),
	);
	const [websocketReady, setWebsocketReady] = useState(false);
	const [isBenchmarking, setIsBenchmarking] = useState(false);

	const submitBenchmark = () => {
		setBenchmarks([]);
		setIsBenchmarking(true);
		websocket.send(`benchmark;${url}`);
	};

	useEffect(() => {
		websocket.addEventListener("open", (event) => {
			console.log("WEBSOCKET OPEN", event);
			setWebsocketReady(true);
		});

		websocket.addEventListener("message", (event) => {
			const [messageType, messageValue] = event.data.split(";");
			switch (messageType) {
				case "url_performance": {
					const benchmark = JSON.parse(messageValue);
					setBenchmarks([...benchmarks, benchmark]);
					break;
				}
				default: {
					console.log("MESSAGE RECEIVED", messageValue);

					if (messageValue === "benchmarking_commplete") {
						setIsBenchmarking(false);
					}
				}
			}
		});

		websocket.addEventListener("error", (event) => {
			console.log("WEBSOCKET ERROR", event);
		});
	}, [websocket, benchmarks]);

	useEffect(() => {
		websocket.addEventListener("close", () => {
			console.log("WEBSOCKET CLOSED, RECONNECTING");
			setWebsocketReady(false);
			setTimeout(() => {
				setWebsocket(new WebSocket("ws://localhost:8000"));
			}, 1000);
		});
	}, [websocket]);

	return (
		<div className="App">
			<div className="logo-container">
				<h1 className="logo">benchmark</h1>
			</div>
			<div className="search-container">
				<input
					type="text"
					placeholder="type url here"
					value={url}
					onChange={(e) => setUrl(e.target.value)}
				/>
				<button
					type="button"
					onClick={submitBenchmark}
					disabled={!websocketReady && !isBenchmarking}
				>
					Press me
				</button>
			</div>
			<div className="results-container">
				{benchmarks.map((benchmark) => {
					return <p key={benchmark.url}>{benchmark.url}</p>;
				})}
			</div>
		</div>
	);
}

export default App;
