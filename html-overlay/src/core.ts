import Promise = require('bluebird');

interface Message {
    type: string;
}

interface MessageCallback {
    (message: Message): void;
}

interface ThrottledCallback {
    (done: {(): void}, ...args: any[]): void
}

interface PendingCall {
    onStart: {(callback: {(): void}): void}
    args: any[]
}

interface ReadyListener {
    (): void;
}

class Core {
    protected _isReady: boolean = false;
    protected _readyListeners: ReadyListener[] = [];
    protected _ws: WebSocket;
    protected _eventListeners: {[id: string]: MessageCallback[]} = {};

    constructor() {
        console.log("Starting Stream-Manager core");
        var core = this;

        var connected = this.connect();
        connected.then(function () {
            console.log("Stream-Manager core is ready.");
            core._isReady = true;
            core._readyListeners.forEach(function (callback) {
                callback();
            });
        }).catch(function (e) {
            console.log("Stream-Manager core failed to initialize.", e);
        });
    }

    public ready(callback: ReadyListener) {
        if (!this._isReady) {
            this._readyListeners.push(callback);
        } else {
            callback();
        }
    }

    public on(messageType: string, callback: MessageCallback) {
        if (this._eventListeners[messageType] === undefined) {
            this._eventListeners[messageType] = [];
        }

        this._eventListeners[messageType].push(callback);
    }

    public throttled(callback: ThrottledCallback) {
        var waiting: PendingCall[] = [];
        var running = false;

        function call(c: PendingCall) {
            var waiterPromise, callPromise;
            waiterPromise = new Promise(function (waiterResolve, reject) {
                callPromise = new Promise(function (resolve, reject) {
                    setTimeout(function () {
                        c.onStart(waiterResolve);
                    }, 1);
                    setTimeout(function () {
                        callback(resolve, ...c.args);
                    }, 0);
                });
            });

            var all = Promise.all([
                callPromise,
                waiterPromise
            ]);

            all.then(function () {
                console.log("Throttled callback ended");
                next();
            });
        }

        function next() {
            var c = waiting.shift();

            if (c) {
                running = true;
                call(c)
            } else {
                running = false;
            }
        }

        return function _throttled(...args: any[]) {
            var p = new Promise(function (resolve, reject) {
                waiting.push({
                    onStart: resolve,
                    args: args,
                });

                if (!running) {
                    next();
                }
            });

            return p;
        }
    }

    protected _handleMessage(message: Message) {
        if (this._eventListeners[message.type] === undefined) {
            console.log(`No handlers for message of type ${message.type}`);
            return;
        }

        this._eventListeners[message.type].forEach(function (callback: MessageCallback) {
            callback(message);
        });
    }

    protected getServerUrl(): string {
        let wsProto = (window.location.protocol === "https:" ? "wss:" : "ws:");
        let base = window.location.origin.replace(window.location.protocol, wsProto);
        return base + "/events";
    }

    protected connect(): Promise {
        var core = this;
        return new Promise(function (resolve, reject) {
            console.log("Connecting to server");
            core._ws = new WebSocket(core.getServerUrl());
            core._ws.onopen = function (event: Event) {
                core._onOpen(event);
                resolve();
            };
            core._ws.onerror = function (event: CloseEvent) {
                try {
                    core._onError(event);
                } catch (e) {
                    reject(e);
                }
            };
            core._ws.onclose = core._onClose.bind(core);
            core._ws.onmessage = core._onMessage.bind(core);
        });
    }

    protected _onOpen(event: Event) {
        console.log("Connected to server.");
    }

    protected _onError(event: Event) {
        throw new Error("Failed to connect to server, check log for details.");
    }

    protected _onClose(event: CloseEvent) {
        console.log("Disconnected from server.");

        var core = this;
        setTimeout(function () {
            core.connect();
        }, 500);
    }

    protected _onMessage(event: MessageEvent) {
        console.log("Got message from server", event);
        var msg: Message = JSON.parse(event.data);
        this._handleMessage(msg);
    }
}

(<any>window).core = new Core();
