package webfriend

//go:generate ./bin/webfriend-autodoc ui/documentation.json
//go:generate esc -o server/static.go -pkg server -modtime 1500000000 -prefix ui ui

const Version = `0.11.4`
const Slogan = `Your friendly friend in web browser automation.`
