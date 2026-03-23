log "Hello"

go "https://news.ycombinator.com"

loop $sub in select "tr.submission .title" {
    log $sub.text
}