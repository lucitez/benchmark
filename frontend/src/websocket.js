export default class CustomWS {
    constructor() {
        this.listeners = []
        this.start()
    }

    start() {
        const sock = new WebSocket("ws://localhost:8000")

        this.sock = sock

        sock.onopen = event => {
            console.log("WEBSOCKET OPEN", event)
        }

        sock.onmessage = (event) => {
            for (const listener of this.listeners) {
                listener(event.data)
            }
        }

        sock.onerror = (event) => {
            console.log("WEBSOCKET ERROR", event)
        }

        sock.onclose = (event) => {
            console.log("WEBSOCKET CLOSED, REOPENING", event)
            setTimeout(() => {
                this.start()
            }, 1000);
        }
    }

    addListener(fn) {
        this.listeners = [...this.listeners, fn]
    }

    removeListener(fn) {
        this.listeners = this.listeners.filter(l => l !== fn)
    }

    send(msg) {
        this.sock.send(msg)
    }
}