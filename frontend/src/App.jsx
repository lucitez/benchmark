import "./App.css";
import { useEffect, useState } from "react";
import BenchmarkResult from "./components/BenchmarkResult";
import CustomWS from "./websocket";

const client = new CustomWS();

function App() {
	const [url, setUrl] = useState("");
	const [benchmarks, setBenchmarks] = useState([]);
	const [status, setStatus] = useState("idle");

	// TODO validate url format
	const submitBenchmark = (e) => {
		e.preventDefault();

		setBenchmarks([]);
		setStatus("idle");
		client.send(`benchmark;${url}`);
	};

	useEffect(() => {
		function handleMessage(message) {
			let [messageType, messageValue] = message.split(";");
			messageValue = messageValue.trim();
			switch (messageType) {
				case "url": {
					setBenchmarks((b) => [...b, { url: messageValue }]);
					break;
				}
				case "status": {
					console.log(messageValue);
					setStatus(messageValue);
					break;
				}
				case "benchmark": {
					const benchmark = JSON.parse(messageValue);
					setBenchmarks((benchmarks) =>
						benchmarks.map((b) => {
							if (b.url === benchmark.url) {
								return benchmark;
							}
							return b;
						}),
					);
					break;
				}
				default: {
					console.log("unhandled message received: ", messageValue);
				}
			}
		}

		client.addListener(handleMessage);

		return () => {
			client.removeListener(handleMessage);
		};
	}, []);

	return (
		<div className="App">
			<div className="logo-container">
				<h1 className="logo">benchmark</h1>
			</div>
			<div className="search-container">
				<form onSubmit={submitBenchmark}>
					<input
						type="text"
						placeholder="enter url"
						value={url}
						onChange={(e) => setUrl(e.target.value)}
					/>
					<button
						disabled={
							url.length === 0 || ["crawling", "benchmarking"].includes(status)
						}
						type="button"
						formAction="submit"
						onClick={submitBenchmark}
					>
						Submit
					</button>
				</form>
			</div>

			<div className="status-container">
				<div className="status">
					{status === "crawling" && "Getting website URLs..."}
					{status === "benchmarking" && "Measuring website performance..."}
					{status === "error" && "Something went wrong, please check your URL"}
					{status === "complete" && "Benchmarking Complete!"}
				</div>
			</div>

			<div className="results-container">
				{benchmarks.map((benchmark) => {
					return (
						<BenchmarkResult
							key={benchmark.url}
							url={benchmark.url}
							latency={benchmark.latency}
						/>
					);
				})}
			</div>
		</div>
	);
}

export default App;
