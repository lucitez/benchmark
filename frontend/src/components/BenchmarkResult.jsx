import React from "react";
import "./BenchmarkResult.css";

function BenchmarkResult(props) {
  return (
    <div className="benchmark-result-container">
      <a
        className="benchmark-url"
        href={props.url}
        target="_blank"
        rel="noreferrer"
      >
        {props.url}
      </a>
      <p>
        {props.latency}
        {props.latency && "ms"}
      </p>
    </div>
  );
}

export default BenchmarkResult;
