'use strict';

window.uuidv4 = function() {
    return 'xxxxxxxxxxxx4xxxyxxxxxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

var DOCTYPE_SVG = '<?xml version="1.0" standalone="no"?>' + '<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">';

var Webfriend = Stapes.subclass({
    constructor: function(container, options){
        this.options = (options || {});
        this.imageStream = null;
        this.commandStream = null;
        this.lastHeader = null;

        this.image = null;
        this.dataSeen = 0;
        this.maxDataSeen = 0;
        this.eventCount = 0;
        this.maxEventCount = 0;
        this.frameCount = 0;
        this.maxFrameCount = 0;
        this.timeOfLastConnectAttempt = null;
        this.negateOffset = 0;

        this.editor = null;
        this.stats = [
            new Stats(),
            new Stats(),
            new Stats(),
            new Stats(),
        ];

        this.statsDownloadPanel = new Stats.Panel('DL', '#F2620C', '#361603');
        this.stats[2].addPanel(this.statsDownloadPanel);

        this.statsEventsPanel = new Stats.Panel('EVT', '#FF4F4C', '#420403');
        this.stats[3].addPanel(this.statsEventsPanel);

        setInterval(function(){
            // events
            var c = this.eventCount;
            this.eventCount = 0;

            if (c > this.maxEventCount) {
                this.maxEventCount = c
            }

            this.statsEventsPanel.update(c, this.maxEventCount);

            var d = this.dataSeen;
            this.dataSeen = 0;

            if (d > this.maxDataSeen) {
                this.maxDataSeen = c
            }

            this.statsDownloadPanel.update(d, this.maxDataSeen);
        }.bind(this), 250);

        this.targetElement = container;

        $(this.targetElement).on('resize', function(){
            this.resizeScreen();
        }.bind(this));

        this.setupStats();
        this.setupKeybindings();
    },

    attachEditor: function(editor) {
        this.editor = editor;
    },

    setupStats: function() {
        if (this.options.statsContainer) {
            $.each(this.stats, function(i, panel){
                panel.dom.style.cssText = 'position: relative';

                if (i <= 1) {
                    panel.showPanel(i);
                } else {
                    panel.showPanel(3);
                }

                $(this.options.statsContainer).append(panel.dom);
            }.bind(this));

            $('#stats').toggle();
        }
    },

    activateInspector: function() {
        this.editor.inspectMode = true;
        this.editor.inspectActive = true;

        $('#inspect').addClass('active');
        console.debug('Activated inspector')
    },

    deactivateInspector: function() {
        this.command('highlight', 'none');
        $('#inspect').removeClass('active');

        this.editor.inspectMode = false;
        this.editor.inspectActive = false;

        this.editor.update();
        console.debug('Deactivated inspector')
    },

    toggleInspectorPause: function(args) {
        this.editor.inspectActive = !this.editor.inspectActive;

        if (this.editor.inspectActive) {
            this.editor.inspectLastPausedOn = null;
            $('#inspect').removeClass('locked');
        } else {
            this.editor.inspectLastPausedOn = args;
            $('#inspect').addClass('locked');
        }
    },

    toggleStats: function() {
        $('#stats').toggle();
    },

    setupKeybindings: function() {
        var lastKnownButton = 0;

        $(this.targetElement).on('contextmenu', function(e) {
            e.preventDefault();

            var parentOffset = $(this.targetElement).offset();
            var relX = e.pageX - parentOffset.left;
            var relY = e.pageY - parentOffset.top;

            this.command('javascript', [
                'return document.elementFromPoint(' + relX.toString() + ',' + relY.toString() +');'
            ], null, false, 'inspect').done(function(reply){
                console.log(reply)
            }.bind(this));
        }.bind(this));

        $(document).on('keydown keyup keypress', function(e) {
            if (this.editor && this.editor.handleEvent(e)) {
                return false;
            } else {
                if (e.target.nodeName === 'BODY') {
                    this.command('key', e.key, {
                        action:  (e.type === 'keyup' ? 'release' : 'press'),
                        alt:     e.altKey,
                        control: e.ctrlKey,
                        meta:    e.metaKey,
                        shift:   e.shiftKey,
                        keycode: e.keyCode,
                    });

                    e.preventDefault();
                }
            }
        }.bind(this));

        $(this.targetElement).on('mousemove mousedown mouseup mousewheel', function(e) {
            try {
                var parentOffset = $(this.targetElement).offset();
                var relX = e.pageX - parentOffset.left;
                var relY = e.pageY - parentOffset.top;

                var args = {
                    x: relX,
                    y: relY,
                };

                var btn = e.button;

                if (e.type == 'mousemove') {
                    btn = lastKnownButton;
                } else {
                    lastKnownButton = btn;
                }

                args['count'] = e.detail;

                switch (btn) {
                case 1:
                    args['button'] = 'middle';
                    break;

                case 2:
                    args['button'] = 'right';
                    break;

                default:
                    args['button'] = 'left';
                    break;
                }

                switch (e.type) {
                case 'mousedown':
                    $('*').blur();
                    args['action'] = 'press';
                    break;

                case 'mouseup':
                    args['action'] = 'release';
                    break;

                case 'mousewheel':
                    args['action'] = 'scroll';
                    args['wheelX'] = -1*e.originalEvent.wheelDeltaX;
                    args['wheelY'] = -1*e.originalEvent.wheelDeltaY;

                    break;
                default:
                    args['action'] = 'move';
                    break;
                }

                if (this.editor && this.editor.inspectMode) {
                    if (args.action === 'press' && args.button === 'left') {
                        this.toggleInspectorPause(args);
                    }

                    if (this.editor.inspectActive) {
                        this.command('inspect', {
                            x: args.x,
                            y: args.y,
                        }).done(function(reply){
                            this.updateInspectNode(reply.scope.result);
                        }.bind(this));
                    } else {
                        args = this.editor.inspectLastPausedOn;

                        this.command('inspect', {
                            x: args.x,
                            y: args.y,
                            r: 242,
                            g: 98,
                            b: 12,
                        });
                    }
                } else {
                    this.command('mouse', args);
                }
            } finally {
                e.preventDefault();
            }
        }.bind(this));
    },

    updateInspectNode: function(node) {
        if (!node) {
            return;
        }

        var inspect = $('#inspect');

        var title = node.name;

        if ($.isPlainObject(node.attributes)) {
            if (node.attributes.id) {
                title += '#' + node.attributes.id;
            } else if (node.attributes.class && node.attributes.class.trim().length) {
                var cls = node.attributes.class.trim();

                title += '.' + cls.replace(/\s+/g, '.');
            }
        }

        inspect.find('.inspect-title').text(title);

        var attributes = $('.inspect-attributes table tbody');

        attributes.empty();

        $.each(node.attributes, function(k, v) {
            var row = $('<tr></tr>');

            row.append($('<th></th>').text(k));
            row.append($('<td></td>').jsonBrowse(v));

            attributes.append(row);
        });
    },

    friendlify: function(arg) {
        if ($.isPlainObject(arg)) {
            var out = '{';

            $.each(arg, function(k, v) {
                out += '\n  ' + k.toString() + ': ' + this.friendlify(v) + ',';
            }.bind(this));

            out += '\n}';

            return out;
        } else {
            switch ($.type(arg)) {
            case 'string':
                if (arg.indexOf('{') >= 0 && arg.indexOf('}') >= 0) {
                    return '"' + arg + '"';
                } else {
                    return "'" + arg + "'";
                }
            }

            return arg.toString();
        }
    },

    command: function(scriptOrCommand, first, rest, tagCommand, resultVar) {
        resultVar = (resultVar || 'result');

        if (!this.commandStream || this.commandStream.readyState != 1) {
            throw 'Command Stream is not available to accept commands';
        }

        if (first || rest) {
            if ($.isPlainObject(first)) {
                scriptOrCommand += ' ' + this.friendlify(first);
            } else if ($.isArray(first)) {
                scriptOrCommand += ' begin\n';
                first.forEach(function(line) {
                    scriptOrCommand += '  ' + line.toString() + '\n';
                });
                scriptOrCommand += '\nend';

            } else {
                scriptOrCommand += ' ' + this.friendlify(first);

                if ($.isPlainObject(rest)) {
                    scriptOrCommand += ' ' + this.friendlify(rest);
                }
            }

            if (scriptOrCommand.indexOf('->') < 0) {
                scriptOrCommand += ' -> $' + resultVar;
            }
        }


        this.deferredReply = $.Deferred(function(){
            // console.debug('SENT')
        }.bind(this));

        if (tagCommand) {
            console.debug('SEND', scriptOrCommand);
            scriptOrCommand = "put '" + uuidv4() + "' -> $invocation;" + scriptOrCommand;
        }

        this.commandStream.send(scriptOrCommand);
        return this.deferredReply;
    },

    connect: function() {
        var now = Date.now();

        if (this.timeOfLastConnectAttempt && (now-this.timeOfLastConnectAttempt) < 750) {
            return;
        }

        this.timeOfLastConnectAttempt = Date.now();

        var p = $.when(
            this.connectImageStream(),
            this.connectCommandStream()
        );

        p.done(function(){
            webfriend.hideOverlay();
        });

        return p;
    },

    disconnect: function(reconnect) {
        if (reconnect) {
            setTimeout(this.connect.bind(this), reconnect);
        }

        return $.when(
            this.stopImageStream(),
            this.stopCommandStream()
        );
    },

    hideOverlay: function() {
        $('#overlay').attr('hidden', 'hidden');
    },

    showOverlay: function() {
        $('#overlay').removeAttr('hidden');
    },

    wsroot: function() {
        var out = '';

        if (window.location.protocol === 'https:') {
            out += 'wss://';
        } else {
            out += 'ws://';
        }

        out += window.location.hostname + ':' + window.location.port;

        return out
    },

    connectImageStream: function() {
        var p = $.Deferred();

        this.imageStream = new WebSocket(
            this.wsroot() + '/api/tabs/current/screencast?q=100&fps=35',
            uuidv4()
        );

        this.imageStream.onerror = function(event) {
            p.reject();
            this.disconnect();
        }.bind(this);

        this.imageStream.onclose = function() {
            this.showOverlay();
            this.disconnect(1000);
        }.bind(this);

        this.imageStream.onopen = function(event) {
            console.debug('Connected to image stream');
            p.resolve();
        }.bind(this);


        this.imageStream.onmessage = function(event) {
            // put the end stats call first; it will no-op on the first frame, but in tandem
            // with the begin call at the end, this will measure the time *between* every
            // subsequent frame.
            try {
                $.each(this.stats, function(i, panel){
                    try {
                        panel.end();
                    } catch(e) {
                        ;
                    }
                });
            } catch (e) {
                ;
            }

            try {
                if ($.type(event.data) === 'string') {
                    if (this.image && this.image.src) {
                        URL.revokeObjectURL(this.image.src);
                    }

                    this.lastHeader = $.parseJSON(event.data);
                    this.image = new Image();

                    this.image.width = this.lastHeader.width;
                    this.image.height = this.lastHeader.height;

                    this.image.onload = function(){
                        var canvas = $(this.targetElement).get(0);
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
            } catch(e) {
                console.error(e)
            } finally {
                $.each(this.stats, function(i, panel){
                    try {
                        panel.begin();
                    } catch(e) {
                        ;
                    }
                });
            }
        }.bind(this);

        return p;
    },

    stopImageStream: function(){
        if (this.imageStream) {
            this.imageStream.close();
            this.imageStream = null;
        }
    },

    connectCommandStream: function() {
        var p = $.Deferred();

        this.commandStream = new WebSocket(
            this.wsroot() + '/api/tabs/current/script',
            uuidv4()
        );

        this.commandStream.onerror = function(event) {
            p.reject();
            this.disconnect();
        }.bind(this);

        this.commandStream.onclose = function() {
            this.showOverlay();
            this.disconnect(1000);
        }.bind(this);

        this.commandStream.onopen = function(event) {
            console.debug('Connected to command stream');
            p.resolve();
        }.bind(this);

        // two kinds of message come back on the command stream: events and replies
        this.commandStream.onmessage = function(event) {
            if (event.data && event.data.length) {
                var reply = $.parseJSON(event.data);

                if (reply.event) {
                    this.processRemoteEvent(reply.event, reply.params);
                } else if (this.deferredReply) {
                    console.log('REPLY', reply)

                    if (reply.success) {
                        this.deferredReply.resolve(reply);
                    } else {
                        this.deferredReply.reject(reply);
                    }
                }
            }
        }.bind(this);

        return p;
    },

    processRemoteEvent: function(name, params) {
        // console.debug('EVENT', name, params);
        this.eventCount += 1;

        switch (name) {
        case 'Webfriend.urlChanged':
            $('#urlbar input[name="url"]').val(params.url);
            break;

        case 'Webfriend.scriptPosted':
            if (this.editor) {
                this.editor.clearMarks();
            }
            break;

        case 'Webfriend.scriptContextEvent':
            if (this.editor && params.id) {
                var severity = 'info';
                var message = '';

                if (params.action === 'finished') {
                    if (params.error) {
                        severity = 'error';
                        message = params.command + ' failed: ' + params.error;
                    } else {
                        severity = 'notice';
                        message = params.command + ' succeeded';
                    }
                }

                if (message.length) {
                    this.editor.log(severity, message, params.took);
                }
            }
            break;

        case 'Network.dataReceived':
            this.dataSeen += params.dataLength;
            break;
        }
    },

    stopCommandStream: function(){
        if (this.commandStream) {
            this.commandStream.close();
            this.commandStream = null;
        }
    },

    resizeScreen: function() {
        var screen = $(this.targetElement).parent();
        var extraY = 0;

        if (screen.width() && screen.height() ) {
            return this.command('resize', {
                width:  screen.width(),
                height: (screen.height() - extraY),
            });
        }
    },
});
