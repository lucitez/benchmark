import "./App.css";
import { useEffect, useState } from "react";
import BenchmarkResult from "./components/BenchmarkResult";

function App() {
	const [url, setUrl] = useState("");
	const [benchmarks, setBenchmarks] = useState([]);
	const [websocket, setWebsocket] = useState();
	const [isBenchmarking, setIsBenchmarking] = useState(false);

	// TODO validate url
	const submitBenchmark = () => {
		setBenchmarks([]);
		setIsBenchmarking(true);
		websocket.send(`benchmark;${url}`);
	};

	useEffect(() => {
		const sock = new WebSocket("ws://localhost:8000");

		sock.addEventListener("open", (event) => {
			console.log("WEBSOCKET OPEN", event);
		});

		sock.addEventListener("error", (event) => {
			console.log("WEBSOCKET ERROR", event);
		});

		// TODO make reconnecting
		sock.addEventListener("error", (event) => {
			console.log("WEBSOCKET CLOSED", event);
		});

		sock.addEventListener("message", (event) => {
			const [messageType, messageValue] = event.data.split(";");
			switch (messageType) {
				case "url_performance": {
					const benchmark = JSON.parse(messageValue);
					setBenchmarks((b) => [...b, benchmark]);
					break;
				}
				default: {
					console.log("MESSAGE RECEIVED", messageValue);
					if (messageValue.trim() === "benchmarking_complete") {
						console.log("OH UEA")
						setIsBenchmarking(false);
					}
				}
			}
		});

		setWebsocket(sock);

		return () => {
			sock.close();
		};
	}, []);

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
				<button disabled={isBenchmarking} type="button" onClick={submitBenchmark}>
					Press me
				</button>
			</div>
			<div className="results-container">
				{benchmarks.map((benchmark) => {
					return <BenchmarkResult key={benchmark.url} url={benchmark.url} />
				})}
			</div>
		</div>
	);
}

export default App;
