'use strict';

var EditorFile = Stapes.subclass({
    constructor: function (buffer, editor) {
        this.id = buffer.id;
        this.editor = editor;
        this.element = d3.select(this.editor.files)
            .append('div')
            .attr('class', 'editor-file')
            .attr('id', 'editor_' + buffer.id)

        this.cm = CodeMirror(document.getElementById('editor_' + buffer.id), {
            mode: 'friendscript',
            theme: 'webfriend',
            indentUnit: 4,
            tabSize: 4,
            lineNumbers: true,
            autofocus: true,
            styleActiveLine: true,
        });

        if (buffer.content) {
            this.cm.setValue(buffer.content);
        }

        this.position = this.cm.cursorCoords(false);
        this.cursor = this.cm.getCursor();

        this.cm.on('cursorActivity', function(cm) {
            this.editor.updateStatusBar(this, cm);
            this.position = cm.cursorCoords(false);
            this.cursor = this.cm.getCursor();
        }.bind(this));

        this.cm.on('changes', function(cm){
            this.persist()
        }.bind(this));
    },

    text: function() {
        return this.cm.getValue();
    },

    persist: function() {
        this.editor.updateBuffer(this.id, {
            'id': this.id,
            'content': this.cm.getValue(),
            'timestamp': Date.now(),
        });
    },

    activate: function() {
        this.element.style('display', 'block');
        this.cm.focus();
        this.cm.scrollIntoView(this.position);
        this.cm.setCursor(this.cursor);
    },
});

var Editor = Stapes.subclass({
    constructor: function (container) {
        this.AUTOSAVE_INTERVAL = 10000;
        this.MIN_AUTOSAVE_INTERVAL_MS = 5000;
        this.SCROLL_FILES_LOOPAROUND = false;

        d3.select(container || 'body').append('div').attr('id', 'editor');

        this.container = '#editor';
        this.editors = [];
        this.activeIndex = null;

        this.buildSkeleton();
        this.loadBuffers();

        if (!this.editors.length) {
            var buffer = this.createFile();
            this.switchToEditor(buffer.id);
        } else {
            this.switchToEditor(this.editors[0].file.id);
        }

        this.setupEvents();
    },

    setupEvents: function() {
        $('.script-execute').on('click', function(e){
            this.executeCurrentBuffer();
            e.preventDefault();
        }.bind(this));
    },

    loadBuffers: function() {
        $.each(this.getAllBuffers(), function(i, buffer) {
            this.createFile(buffer);
        }.bind(this));
    },

    createFile: function (buffer) {
        if (!buffer) {
            buffer = {
                'id': Date.now().toString(),
                'content': '',
            };
        }

        this.updateBuffer(buffer.id, buffer);
        buffer = this.retrieveBuffer(buffer.id);

        this.editors.push({
            'editor': this,
            'file': new EditorFile(buffer, this),
        });

        this.update();
        this.switchToEditor(buffer.id);

        return buffer;
    },

    buildSkeleton: function () {
        // Filebar
        // -----------------------------------------------------------------------------------------
        var filebar = d3.select(this.container)
            .append('div')
            .attr('class', 'filebar');

        filebar
            .append('ul')
            .attr('class', 'nav nav-tabs workspace-selector');

        var actions = filebar
            .append('ul')
            .attr('class', 'nav nav-tabs file-actions');

        actions
            .append('li').attr('class', 'nav-item')
            .append('a').attr('class', 'nav-link new-file')
            .on('click', function(){
                this.createFile();
            }.bind(this))
            .append('i').attr('class', 'fa fa-fw fa-plus')

        filebar
            .append('ul')
            .attr('class', 'nav nav-tabs open-files');

        // Files View
        // -----------------------------------------------------------------------------------------
        d3.select(this.container)
            .append('div')
            .attr('class', 'files');

        // Statusbar
        // -----------------------------------------------------------------------------------------
        var sb = d3.select(this.container)
            .append('div')
            .attr('class', 'statusbar')
            .append('div')
            .attr('class', 'statusbar-cursor');

        sb.append('span').text('Line ');
        sb.append('span').attr('class', 'line').text('-');
        sb.append('span').text(',Col ');
        sb.append('span').attr('class', 'ch').text('-');

        this.filebar = '#editor > .filebar';
        this.files = '#editor > .files';
        this.statusbar = '#editor > .statusbar';

        // $(window).on('click', function(e){
        //     var el = $(e.target);

        //     $.each($(this.container + ' .open-files .nav-item'), function(i, item){
        //         if ($.contains(item, e.target)) {
        //             var a = el.closest('a');
        //             this.switchToEditor(a.attr('data-editor-buffer'));
        //         }
        //     }.bind(this));

        // }.bind(this));

    },

    updateFilebar: function () {
        var files = d3.select(this.filebar + ' .open-files')
            .selectAll('li')
            .data(this.editors);

        // BEGIN enter selection
        var tab = files.enter()
            .append('li')
            .attr('class', 'nav-item')

        var fileLink = tab.append('a')
            .attr('href', '#')
            .on('click', function(d) {
                var el = $(this);
                var id = el.parent().attr('data-editor-id');

                if (id) {
                    console.debug('clicked on', id, d.file.id);
                    d.editor.switchToEditor(id);
                }
            });

        fileLink.append('i')
            .attr('class', 'fa fa-fw fa-hashtag file-icon')

        fileLink
            .append('span')
            .attr('class', 'filename')
            .text(function (d) {
                return (d.file.filename || 'Untitled');
            });

        tab.append('a')
            .attr('href', '#')
            .attr('data-action', 'close-file')
            .on('click', function(d) {
                var el = $(this);
                var id = el.parent().attr('data-editor-id');

                if (id) {
                    d.editor.remove(id);
                }
            })
            .append('i')
            .attr('class', 'fa fa-fw fa-times');

        // END enter selection

        tab.merge(files)
            .attr('data-editor-id', function(d){
                return d.file.id;
            })
            .selectAll('a')
            .attr('class', function (d) {
                if (this.activeIndex == d.file.id) {
                    return 'nav-link active';
                } else {
                    return 'nav-link';
                }
            }.bind(this));

        files.exit()
            .transition()
            .duration(100)
            .ease(d3.easeLinear)
            .style('opacity', 0)
            .remove();
    },

    updateVisibleEditor: function(editor) {
        if (editor) {
            $('#editor .files > .editor-file').css('display', 'none');
            editor.file.activate();
        }
    },

    switchToEditor: function (id) {
        if (id) {
            this.activeIndex = id;
            this.update();
        }
    },

    getEditorByIndex: function(id) {
        var out = null;

        $.each(this.editors, function(i, editor){
            if (editor.file.id == id) {
                out = editor;
            }
        });

        return out;
    },

    update: function() {
        this.updateFilebar();

        var editor = this.getEditorByIndex(this.activeIndex);

        if (editor) {
            this.updateVisibleEditor(editor);
            this.updateStatusBar(editor.file, editor.file.cm);
        } else if (this.editors.length === 0) {
            $('#editor .files').empty();
        }
    },

    updateStatusBar: function (file, cm) {
        if (this.activeIndex == file.id) {
            var sb = $('#editor .statusbar');
            var cursor = cm.getCursor();

            sb.find('.line').text(cursor.line + 1);
            sb.find('.ch').text(cursor.ch + 1);
        }
    },


    updateBuffer: function (id, document) {
        if (window.localStorage) {
            localStorage.setItem('webfriend.editor.buffer.' + id, JSON.stringify(document));
            console.debug('Saved buffer', id);
        } else {
            throw 'Local Storage is not available.';
        }
    },

    retrieveBuffer: function (id) {
        if (window.localStorage) {
            var document = localStorage.getItem('webfriend.editor.buffer.' + id);

            if (document) {
                return JSON.parse(document);
            } else {
                return {};
            }
        } else {
            throw 'Local Storage is not available.';
        }
    },

    getAllBuffers: function () {
        var buffers = [];

        $.each(Object.keys(localStorage), function (i, key) {
            key = key.replace(/^webfriend\.editor\.buffer\./, '');

            var buffer = this.retrieveBuffer(key);
            buffers.push(buffer);
        }.bind(this));

        return buffers;
    },

    remove: function(id) {
        localStorage.removeItem('webfriend.editor.buffer.' + id);

        this.editors = $.grep(this.editors, function(v){
            return v.file.id != id;
        }.bind(this));

        this.update();
    },

    executeCurrentBuffer: function() {
        if (this.activeIndex) {
            console.debug('Execute buffer', this.activeIndex);

            var editor = this.getEditorByIndex(this.activeIndex);

            if (editor) {
                webfriend.command(editor.file.text()).done(function(reply){
                    console.log(reply);
                }.bind(this)).fail(function(reply){
                    console.error(reply);
                }.bind(this))
            }
        }
    },
});