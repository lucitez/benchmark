import "./App.css";
import { useEffect, useState } from "react";
import BenchmarkResult from "./components/BenchmarkResult";
import CustomWS, { Message } from "./websocket";
import type { Benchmark } from "./types";

const client = new CustomWS();

type status = "idle" | "crawling" | "benchmarking" | "complete" | "error";

function App() {
	const [url, setUrl] = useState("");
	const [benchmarks, setBenchmarks] = useState<Benchmark[]>([]);
	const [numUrlsBenchmarked, setNumUrlsBenchmarked] = useState(0);
	const [status, setStatus] = useState<status>("idle");
	const [error, setError] = useState("");

	// TODO validate url format
	const startBenchmark = () => {
		const expression =
			/[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&//=]*)/gi;
		const regex = new RegExp(expression);

		if (!url.match(regex)) {
			setStatus("error");
			setError("Please enter a valid url");
			return;
		}

		setNumUrlsBenchmarked(0);
		setBenchmarks([]);
		setStatus("idle");
		setError("");
		client.send(`benchmark;${url}`);
	};

	useEffect(() => {
		document.title = "benchmark - measure your website's performance";
	}, []);

	useEffect(() => {
		function handleMessage({ type, value }: Message) {
			switch (type) {
				case "url": {
					setBenchmarks((b) => [...b, { url: value }]);
					break;
				}
				case "status": {
					console.log(value);
					setStatus(value as status);
					if (value === "complete") {
						setTimeout(() => {
							setStatus("idle");
						}, 2000);
					}
					break;
				}
				case "benchmark": {
					const benchmark = JSON.parse(value);
					setBenchmarks((benchmarks) =>
						benchmarks.map((b) => {
							if (b.url === benchmark.url) {
								return benchmark;
							}
							return b;
						}),
					);
					setNumUrlsBenchmarked((n) => n + 1);
					break;
				}
				default: {
					console.log("unhandled message received: ", value);
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
			<div className="subtitle-container">
				<div className="subtitle">
					enter a url to measure a website's performance
				</div>
			</div>

			<form
				className="search-form"
				onSubmit={(e) => {
					e.preventDefault();
					startBenchmark();
				}}
			>
				<input
					className="url-input"
					type="text"
					placeholder="enter url"
					value={url}
					onChange={(e) => setUrl(e.target.value)}
				/>
				<button
					className="url-submit"
					disabled={
						url.length === 0 || ["crawling", "benchmarking"].includes(status)
					}
					type="button"
					formAction="submit"
					onClick={startBenchmark}
				>
					start
				</button>
			</form>

			<div className="status-container">
				<div className="status">
					{status === "crawling" &&
						`crawling ${url} URLs: ${benchmarks.length}`}
					{status === "benchmarking" &&
						`measuring performance: (${numUrlsBenchmarked}/${benchmarks.length})`}
					{status === "error" &&
						(error || "Something went wrong, please check your URL")}
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
