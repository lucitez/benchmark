import "./BenchmarkResult.css";

type Props = {
	url: string;
	latency: number | undefined;
};

function BenchmarkResult({ url, latency }: Props) {
	return (
		<div className="benchmark-result-container">
			<a className="benchmark-url" href={url} target="_blank" rel="noreferrer">
				{url}
			</a>
			<p>
				{latency}
				{latency && "ms"}
			</p>
		</div>
	);
}

export default BenchmarkResult;
