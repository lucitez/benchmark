export type Message = {
    type: string;
    value: string
}
type Listener = (msg: Message) => void

export default class CustomWS {
    private listeners: Listener[]
    private sock!: WebSocket;

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
            let [type, value] = event.data.split(";");
            value = value.trim();
            for (const listener of this.listeners) {
                listener({ type, value })
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

    addListener(listener: Listener) {
        this.listeners = [...this.listeners, listener]
    }

    removeListener(listener: Listener) {
        this.listeners = this.listeners.filter(l => l !== listener)
    }

    send(msg: string) {
        this.sock.send(msg)
    }
}