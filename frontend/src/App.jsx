import "./App.css";
import { useEffect, useState } from "react";
import BenchmarkResult from "./components/BenchmarkResult";
import CustomWS from './websocket'

const client = new CustomWS()

function App() {
	const [url, setUrl] = useState("");
	const [benchmarks, setBenchmarks] = useState([]);
	const [isBenchmarking, setIsBenchmarking] = useState(false);

	// TODO validate url format
	const submitBenchmark = () => {
		setBenchmarks([]);
		setIsBenchmarking(true);
		client.send(`benchmark;${url}`);
	};

	useEffect(() => {
		function handleMessage(message) {
			const [messageType, messageValue] = message.split(";");
			switch (messageType) {
				case "url_performance": {
					const benchmark = JSON.parse(messageValue);
					setBenchmarks((b) => [...b, benchmark]);
					break;
				}
				default: {
					console.log("MESSAGE RECEIVED", messageValue);
					if (messageValue.trim() === "benchmarking_complete") {
						setIsBenchmarking(false);
					}
				}
			}
		}

		client.addListener(handleMessage)

		return () => {
			client.removeListener(handleMessage)
		}
	}, [])

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
					return <BenchmarkResult key={benchmark.url} url={benchmark.url} latency={benchmark.latency} />
				})}
			</div>
		</div>
	);
}

export default App;
