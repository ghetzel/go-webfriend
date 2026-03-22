window.uuidv4 = function () {
    return 'xxxxxxxxxxxx4xxxyxxxxxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

window.qs = function (key, fallback) {
    return (new URLSearchParams(window.location.search)).get(key) || fallback
}

Vue.createApp({
    delimiters: ['[[', ']]'],

    data: () => ({
        api: '/api',
        intendedUrl: function () {
            var url = qs('url', localStorage.getItem('url'))
            if (!url) {
                url = location.protocol + '//' + location.host + '/home'
            }
            return url
        }(),
        imageStream: null,
        targetElement: '#browser',
        bytesReceived: 0,
        eventCount: 0,
        imageQuality: parseInt(qs('q', 100)),
        imageFPS: parseInt(qs('fps', 60)),
    }),

    created() {
        // window.onbeforeunload = this.disconnect

        window.addEventListener('resize', function () {
            this.resizeScreen();
        }.bind(this));

        this.connect().then(function () {
            this.resizeScreen()
            this.navigate()
        }.bind(this))

        this.setupPostConnectBindings()
    },

    watch: {
        intendedUrl: "navigate",
        imageQuality: "reconnectImageStream",
        imageFPS: "reconnectImageStream",
    },

    methods: {
        // -------------------------------------------------------------------------------------------------- //
        async httpGet(path, discard) {
            var res = await fetch(`${this.api}${path}`)

            if (!discard) {
                return await res.json()
            }
        },
        // -------------------------------------------------------------------------------------------------- //
        navigate(url) {
            url = url || this.intendedUrl
            localStorage.setItem("url", url)
            return this.command('go', url)
        },
        // -------------------------------------------------------------------------------------------------- //
        connect() {
            var now = Date.now();

            if (this.timeOfLastConnectAttempt && (now - this.timeOfLastConnectAttempt) < 750) {
                return;
            }

            this.timeOfLastConnectAttempt = Date.now();

            return Promise.all([
                this.connectImageStream(),
                this.connectCommandStream(),
            ])
        },
        // -------------------------------------------------------------------------------------------------- //
        disconnect(reconnect) {
            if (reconnect) {
                setTimeout(this.connect.bind(this), reconnect);
            }

            return Promise.all([
                this.stopImageStream(),
                this.stopCommandStream(),
            ])
        },
        // -------------------------------------------------------------------------------------------------- //
        wsroot() {
            var out = '';

            if (window.location.protocol === 'https:') {
                out += 'wss://';
            } else {
                out += 'ws://';
            }

            out += window.location.hostname + ':' + window.location.port;

            return out
        },
        // -------------------------------------------------------------------------------------------------- //
        imageQuality() {
            this.stopImageStream()
            this.connectImageStream()
        },
        // -------------------------------------------------------------------------------------------------- //
        connectImageStream() {
            return new Promise((resolve, reject) => {
                this.imageStream = new WebSocket(
                    this.wsroot() + `/api/tabs/current/screencast?q=${this.imageQuality}&fps=${this.imageFPS}`,
                    uuidv4()
                );

                this.imageStream.onerror = function (event) {
                    reject()
                    this.disconnect()
                }.bind(this)

                this.imageStream.onclose = function () {
                    // this.showOverlay()
                    this.disconnect(1000)
                }.bind(this)

                this.imageStream.onopen = function (event) {
                    console.debug('Connected to image stream')
                    resolve()
                }.bind(this)


                this.imageStream.onmessage = function (event) {
                    // put the end stats call first; it will no-op on the first frame, but in tandem
                    // with the begin call at the end, this will measure the time *between* every
                    // subsequent frame.
                    try {
                        this.stats.forEach((panel) => panel.end())
                    } catch (e) {
                        ;
                    }

                    try {
                        if (typeof event.data == 'string') {
                            if (this.image && this.image.src) {
                                URL.revokeObjectURL(this.image.src);
                            }

                            this.lastHeader = JSON.parse(event.data);
                            this.image = new Image();

                            this.image.width = this.lastHeader.width;
                            this.image.height = this.lastHeader.height;

                            this.image.onload = function () {
                                var canvas = document.querySelector(this.targetElement);

                                if (!canvas) {
                                    return
                                }

                                var ctx = canvas.getContext('2d');

                                ctx.clearRect(0, 0, canvas.width, canvas.height);

                                if (canvas.width != this.image.width) {
                                    canvas.width = this.image.width;
                                }

                                if (canvas.height != this.image.height) {
                                    canvas.height = this.image.height;
                                }

                                ctx.drawImage(this.image, 0, 0);
                            }.bind(this);

                        } else if (this.lastHeader && this.image) {
                            this.image.src = URL.createObjectURL(event.data);
                        }
                    } catch (e) {
                        console.error(e)
                    } finally {
                        try {
                            this.stats.forEach((panel) => panel.begin())
                        } catch (e) {
                            ;
                        }
                    }
                }.bind(this);
            })
        },
        // -------------------------------------------------------------------------------------------------- //
        stopImageStream() {
            if (this.imageStream) {
                this.imageStream.close();
                this.imageStream = null;
            }
        },
        // -------------------------------------------------------------------------------------------------- //
        connectCommandStream() {
            return new Promise((resolve, reject) => {
                this.commandStream = new WebSocket(
                    this.wsroot() + '/api/tabs/current/script',
                    uuidv4()
                );

                this.commandStream.onerror = function (event) {
                    reject()
                    this.disconnect()
                }.bind(this)

                this.commandStream.onclose = function () {
                    // this.showOverlay()
                    this.disconnect(1000)
                }.bind(this)

                this.commandStream.onopen = function (event) {
                    console.debug('Connected to command stream')
                    resolve()
                }.bind(this)

                // two kinds of message come back on the command stream: events and replies
                this.commandStream.onmessage = function (event) {
                    if (event.data && event.data.length) {
                        var reply = JSON.parse(event.data)

                        if (reply.event) {
                            this.processRemoteEvent(reply.event, reply.params)
                        } else if (this.deferredReply) {
                            if (reply.success) {
                                this.deferredReply.then((resolve) => resolve(reply))
                            } else {
                                this.deferredReply.then((_, reject) => reject(reply))
                            }
                        }
                    }
                }.bind(this)
            })
        },
        // -------------------------------------------------------------------------------------------------- //
        stopCommandStream() {
            if (this.commandStream) {
                this.commandStream.close()
                this.commandStream = null
            }
        },
        // -------------------------------------------------------------------------------------------------- //
        processRemoteEvent(name, params) {
            // console.debug('EVENT', name, params);
            this.eventCount += 1;

            switch (name) {
                case 'Webfriend.urlChanged':
                    this.currentUrl = params.url
                    break;
                case 'Webfriend.scriptPosted':
                    if (this.editor) {
                        this.editor.clearMarks()
                    }
                    break

                case 'Webfriend.scriptContextEvent':
                    if (this.editor && params.id) {
                        var severity = 'info'
                        var message = ''

                        if (params.action === 'finished') {
                            if (params.error) {
                                severity = 'error'
                                message = params.command + ' failed: ' + params.error
                            } else {
                                severity = 'notice'
                                message = params.command + ' succeeded'
                            }
                        }

                        if (message.length) {
                            this.editor.log(severity, message, params.took)
                        }
                    }
                    break

                case 'Network.dataReceived':
                    this.bytesReceived += params.dataLength
                    break
            }
        },
        // -------------------------------------------------------------------------------------------------- //
        resizeScreen() {
            var screen = document.querySelector('body').getBoundingClientRect();
            var extraY = 0;

            if (screen.width && screen.height) {
                return this.command('resize', {
                    width: screen.width,
                    height: (screen.height - extraY),
                });
            }
        },
        // -------------------------------------------------------------------------------------------------- //
        friendlify(arg) {
            if (typeof arg == 'object') {
                var out = '{'

                Object.keys(arg).forEach(function (k) {
                    out += '\n  ' + k.toString() + ': ' + this.friendlify(arg[k]) + ','
                }.bind(this))

                out += '\n}'

                return out
            } else {
                switch (typeof arg) {
                    case 'string':
                        if (arg.indexOf('{') >= 0 && arg.indexOf('}') >= 0) {
                            return '"' + arg + '"'
                        } else {
                            return "'" + arg + "'"
                        }
                }

                return arg.toString()
            }
        },
        // -------------------------------------------------------------------------------------------------- //
        command(scriptOrCommand, first, rest, tagCommand, resultVar) {
            resultVar = (resultVar || 'result');

            if (!this.commandStream || this.commandStream.readyState != 1) {
                throw 'Command Stream is not available to accept commands';
            }

            if (first || rest) {
                if (typeof first == 'object') {
                    scriptOrCommand += ' ' + this.friendlify(first);
                } else if (Array.isArray(first)) {
                    scriptOrCommand += ' begin\n';
                    first.forEach((line) => {
                        scriptOrCommand += '  ' + line.toString() + '\n';
                    });
                    scriptOrCommand += '\nend';

                } else {
                    scriptOrCommand += ' ' + this.friendlify(first);

                    if (typeof rest == 'object') {
                        scriptOrCommand += ' ' + this.friendlify(rest);
                    }
                }

                if (scriptOrCommand.indexOf('->') < 0) {
                    scriptOrCommand += ' -> $' + resultVar;
                }
            }


            this.deferredReply = new Promise(function (resolve, reject) {
                ;
            }.bind(this));

            if (tagCommand) {
                console.debug('SEND', scriptOrCommand);
                scriptOrCommand = "put '" + uuidv4() + "' -> $invocation;" + scriptOrCommand;
            }

            this.commandStream.send(scriptOrCommand);
            return this.deferredReply;
        },
        // -------------------------------------------------------------------------------------------------- //
        setupPostConnectBindings() {
            // window.addEventListener('contextmenu', function (e) {
            //     e.preventDefault();

            //     // var el = document.querySelector(this.targetElement)
            //     // var relX = e.pageX - el.offsetLeft;
            //     // var relY = e.pageY - el.offsetTop;

            //     // this.command('javascript', [
            //     //     'return document.elementFromPoint(' + relX.toString() + ',' + relY.toString() + ');'
            //     // ], null, false, 'inspect');
            // }.bind(this));

            window.addEventListener('mousemove', this.handleMouseEvent.bind(this))
            window.addEventListener('mousedown', this.handleMouseEvent.bind(this))
            window.addEventListener('mouseup', this.handleMouseEvent.bind(this))
            window.addEventListener('mousewheel', this.handleMouseEvent.bind(this))
        },
        // -------------------------------------------------------------------------------------------------- //
        handleMouseEvent(e) {
            var lastKnownButton = 0;

            try {
                var parentOffset = document.querySelector(this.targetElement).offsetParent
                var relX = e.pageX - parentOffset.offsetLeft
                var relY = e.pageY - parentOffset.offsetTop

                var args = {
                    x: relX,
                    y: relY,
                }

                var btn = e.button

                if (e.type == 'mousemove') {
                    btn = lastKnownButton
                } else {
                    lastKnownButton = btn
                }

                args['count'] = e.detail

                switch (btn) {
                    case 1:
                        args['button'] = 'middle'
                        break

                    case 2:
                        args['button'] = 'right'
                        break

                    default:
                        args['button'] = 'left'
                        break
                }

                switch (e.type) {
                    case 'mousedown':
                        args['action'] = 'press'
                        break

                    case 'mouseup':
                        args['action'] = 'release'
                        break

                    case 'mousewheel':
                        args['action'] = 'scroll'
                        args['wheelX'] = -1 * e.originalEvent.wheelDeltaX
                        args['wheelY'] = -1 * e.originalEvent.wheelDeltaY

                        break
                    default:
                        args['action'] = 'move'
                        break
                }

                // this.command('inspect', {
                //     x: args.x,
                //     y: args.y,
                //     r: 242,
                //     g: 98,
                //     b: 12,
                // })

                this.command('mouse', args)
            } catch (e) {
                console.error(e)
            } finally {
                e.preventDefault()
            }
        },
    },
}).mount('main')