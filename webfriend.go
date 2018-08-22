package webfriend

//go:generate ./bin/webfriend-autodoc ui/documentation.gob
//go:generate esc -o server/static.go -pkg server -modtime 1500000000 -prefix ui ui

const Version = `0.9.18`
const Slogan = `Your friendly friend in web browser automation.`
