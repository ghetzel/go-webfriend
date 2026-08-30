#!/usr/bin/env webfriend
log "[[208:237]]HackerNews                                                                      [[reset]]"

go "https://news.ycombinator.com"

loop $sub in selectAll "tr.submission .title" {
    log "[[208:237]]{sub.text}[[reset]]"
}