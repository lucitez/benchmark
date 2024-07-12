import "./App.css";
import { useEffect, useState } from "react";
import BenchmarkResult from "./components/BenchmarkResult";
import CustomWS, { Message } from "./websocket";
import { Benchmark } from "./types";

const client = new CustomWS();

type status = "idle" | "crawling" | "benchmarking" | "complete" | "error";

function App() {
  const [url, setUrl] = useState("");
  const [benchmarks, setBenchmarks] = useState<Benchmark[]>([]);
  const [numUrlsBenchmarked, setNumUrlsBenchmarked] = useState(0);
  const [status, setStatus] = useState<status>("idle");

  // TODO validate url format
  const startBenchmark = () => {
    setNumUrlsBenchmarked(0);
    setBenchmarks([]);
    setStatus("idle");
    client.send(`benchmark;${url}`);
  };

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
            })
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
      <div className="search-container">
        <form
          onSubmit={(e) => {
            e.preventDefault();
            startBenchmark();
          }}
        >
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
            onClick={startBenchmark}
          >
            Start
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

      {["crawling", "benchmarking"].includes(status) && (
        <div className="progress-container">
          <div className="progress">
            {status === "crawling" && `URLs found: ${benchmarks.length}`}
            {status === "benchmarking" &&
              `Progress: (${numUrlsBenchmarked}/${benchmarks.length})`}
          </div>
        </div>
      )}

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
