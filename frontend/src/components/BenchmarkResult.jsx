import React from "react";
import "./BenchmarkResult.css";

function BenchmarkResult(props) {
	return (
		<div className="benchmark-result-container">
			<p>{props.url}</p>
			<p>{props.latency}</p>
		</div>
	);
}

export default BenchmarkResult;
