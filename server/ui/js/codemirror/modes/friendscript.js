"use strict";
CodeMirror.defineMode('friendscript', function (cm, options) {
    var IDENT = /[a-z_A-Z0-9\.]/;
    var NUMERIC = /\d+/;
    var SPACE = /\s+/;
    var KEYWORDS = [
        'if',
        'else',
        'in',
        'not',
        'break',
        'continue',
        'loop',
        'count',
    ];

    var BLOCK_OPENERS = [
        'loop',
        'if',
        'else',
    ];

    var ATOMS = [
        'true',
        'false',
        'null',
    ];

    var line = 0;

    return {
        startState: function () {
            return {
                lastMode:     null,
                inBlock:      false,
                inObject:     0,
                inString:     false,
                inVariable:   false,
                lhs:          false,
                stringType:   '',
                lastModule:   '',
                lastCommand:  '',
                lastTag:      '',
                prefixAtOpen: '',
            };
        },

        electricInput: /\s{2,}$/,

        indent: function(state, textAfter) {
            if (textAfter == '}') {
                return (cm.indentUnit * (state.inObject - 1));
            } else {
                return (cm.indentUnit * state.inObject);
            }
        },

        token: function (stream, state) {
            if (stream.sol()) {
                line += 1;
                state.lastTag = '';
            }

            // we're about to start a string
            if (!state.inString && (stream.peek() == '"' || stream.peek() == "'")) {
                state.stringType = stream.peek();
                stream.next(); // Skip quote
                state.inString = true; // Update state
            }

            if (stream.sol()) {
                state.lhs = true;
            }

            if (state.inString) {
                if (stream.match(/^end/)) {
                    state.inString = false;
                    return 'keyword';
                } else {
                    while (state.inString && !stream.eol()) {
                        if (stream.peek() === state.stringType) {
                            stream.next(); // Skip quote
                            state.inString = false; // Clear flag

                            if (stream.peek() == ']') {
                                state.inVariable = true;
                            }
                        } else if (stream.peek() === '\\') {
                            stream.next();
                            stream.next();
                        } else {
                            stream.match(/^.[^\\\"\']*/);
                        }
                    }

                    return 'string';
                }
            } else if (stream.peek() === '#') {
                stream.skipToEnd();
                return 'comment';
            } else if (!state.inBlock && stream.peek() == '{') {
                state.inObject += 1;
                stream.next();
                return 'bracket';

            } else if (stream.peek() == '}') {
                if (state.inObject) {
                    state.inObject -= 1;
                } else if (state.inBlock) {
                    state.inBlock = false;
                }

                stream.next();
                return 'bracket';

            } else if (state.lhs && stream.peek() === '=') {
                state.lhs = false;

            } else if (stream.peek() === '$') {
                state.inVariable = true;
                stream.next();
                stream.eatWhile(IDENT);
                return 'variable';

            } else if (state.inVariable && (stream.peek() === '[' || stream.sol())) {
                state.inVariable = false;

            } else if (!state.inVariable && stream.peek() === ']') {
                stream.next();

                if (stream.peek() == '.') {
                    state.inVariable = true;
                }

                return null;

            } else if (state.inVariable) {
                stream.eatWhile(IDENT);

                if (stream.peek() != '.') {
                    state.inVariable = false;
                }

                return 'variable';
            }

            // tokenize & identify words
            // -------------------------------------------------------------------------------------
            if (IDENT.test(stream.peek())) {
                var word = ''

                stream.eatWhile(function(c) {
                    if (IDENT.test(c)) {
                        word += c;
                        return true;
                    } else {
                        return false;
                    }
                });

                if (KEYWORDS.indexOf(word) >= 0) {
                    if (BLOCK_OPENERS.indexOf(word) >= 0) {
                        state.inBlock = true;
                    }

                    return 'keyword';
                } else if (NUMERIC.test(word)) {
                    return 'number';
                } else if (ATOMS.indexOf(word) >= 0) {
                    return 'atom';
                } else if (!state.inString && word == 'begin') {
                    state.inString = true;
                    return 'keyword';
                } else if (state.inObject) {
                    state.lastTag = word;
                    return 'tag';

                } else if (!state.inObject) {
                    if (stream.peek() === ':') {
                        state.lastModule = word;
                    } else {
                        state.lastCommand = word;
                    }

                    return 'property';
                }
            }

            stream.next();
            return null;
        },
    };
});