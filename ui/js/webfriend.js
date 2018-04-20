'use strict';

window.uuidv4 = function() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

var Webfriend = Stapes.subclass({
    constructor: function(options){
        this.options = (options || {});
        this.imageStream = null;
        this.commandStream = null;
        this.lastHeader = null;
        this.imageStats = [ new Stats(), new Stats() ];
        this.targetElementId = '#browser';
    },

    setScreenTarget: function(targetElementId) {
        this.targetElementId = targetElementId;
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

    command: function(scriptOrCommand, first, rest) {
        if (!this.commandStream || this.commandStream.readyState != 1) {
            throw 'Command Stream is not available to accept commands';
        }

        if (first || rest) {
            if ($.isPlainObject(first)) {
                scriptOrCommand += ' ' + this.friendlify(first);
            } else {
                scriptOrCommand += ' ' + this.friendlify(first);

                if ($.isPlainObject(rest)) {
                    scriptOrCommand += ' ' + this.friendlify(rest);
                }
            }
        }

        // console.debug('SEND', scriptOrCommand);

        this.deferredReply = $.Deferred(function(){
            // console.debug('SENT')
        }.bind(this));
        this.commandStream.send(scriptOrCommand);
        return this.deferredReply;
    },

    connect: function() {
        return $.when(
            this.connectImageStream(),
            this.connectCommandStream()
        );
    },

    disconnect: function() {
        this.stopImageStream();
        this.stopCommandStream();
    },

    wsroot: function() {
        var out = '';

        if (window.location.scheme === 'https') {
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
            this.stopImageStream();
            this.connectImageStream();
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
                $.each(this.imageStats, function(i, panel){
                    panel.end();
                });
            } catch (e) {
                ;
            }

            try {
                if ($.type(event.data) === 'string') {
                    this.lastHeader = $.parseJSON(event.data);
                    var img = $(this.targetElementId);

                    img.css('width', this.lastHeader.width);
                    img.css('height', this.lastHeader.height);
                } else if (this.lastHeader) {
                    var img = $(this.targetElementId).get(0);
                    img.src = URL.createObjectURL(event.data);
                }
            } catch(e) {
                console.error(e)
            } finally {
                $.each(this.imageStats, function(i, panel){
                    panel.begin();
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
            this.stopCommandStream();
            this.connectCommandStream();
        }.bind(this);

        this.commandStream.onopen = function(event) {
            console.debug('Connected to command stream');
            p.resolve();
        }.bind(this);

        this.commandStream.onmessage = function(event) {
            var reply = $.parseJSON(event.data);

            if (this.deferredReply) {
                if (reply.success) {
                    this.deferredReply.resolve(reply);
                } else {
                    this.deferredReply.reject(reply);
                }
            }
        }.bind(this);

        return p;
    },

    stopCommandStream: function(){
        if (this.commandStream) {
            this.commandStream.close();
            this.commandStream = null;
        }
    },

    resizeScreen: function() {
        var screen = $(this.targetElementId).parent();

        if (screen.width() && screen.height() ) {
            return this.command('resize', {
                width:  screen.width(),
                height: screen.height(),
            });
        }
    },
});

$(document).ready(function(){
    window.webfriend = new Webfriend();
});