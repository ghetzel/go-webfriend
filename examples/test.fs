log "Hello"

go "https://www.humblebeeandme.com/project-sitemap.xml" {
    load_event_name: 'idle',
}

log $result

loop $sub in select "loc" {
    log $sub.text
}