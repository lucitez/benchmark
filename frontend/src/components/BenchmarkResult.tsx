import type { Benchmark } from "../types";
import "./BenchmarkResult.css";

type Props = {
	benchmark: Benchmark;
};

function BenchmarkResult({ benchmark }: Props) {
	const { url, latency, size } = benchmark;

	let shortenedSize = size;
	let sizeSuffix = "b";
	if (size && size > 999) {
		shortenedSize = size / 1000;
		sizeSuffix = "kb";
	}

	return (
		<div className="benchmark-result-container">
			<a className="benchmark-url" href={url} target="_blank" rel="noreferrer">
				{url}
			</a>
			<div className="stats-container">
				<p className="size">
					{shortenedSize?.toFixed(2)}
					{shortenedSize && sizeSuffix}
				</p>
				<p>
					{latency}
					{latency && "ms"}
				</p>
			</div>
		</div>
	);
}

export default BenchmarkResult;
