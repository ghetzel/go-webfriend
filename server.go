package webfriend

//go:generate esc -o static.go -pkg webfriend -modtime 1500000000 -prefix ui ui

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ghetzel/diecast"
	"github.com/ghetzel/go-stockutil/httputil"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/gorilla/websocket"
	"github.com/husobee/vestigo"
	"github.com/urfave/negroni"
)

type ClientSessionType int

const (
	ImageSession ClientSessionType = iota
	CommandSession
)

var SessionConnCheckInterval = 10 * time.Second
var GlobalSessionPollInterval = 32 * time.Millisecond

type clientSession struct {
	Type            ClientSessionType
	Tab             *browser.Tab
	Session         string
	Conn            *websocket.Conn
	Server          *Server
	InterframeDelay time.Duration
	LastFrameTime   time.Time
	lastFrameW      int
	lastFrameH      int
	lastFrameId     int64
	waiterId        string
}

func (self *clientSession) prep() {
	go func() {
		pinger := time.NewTicker(SessionConnCheckInterval)
		defer pinger.Stop()

		for {
			select {
			case <-pinger.C:
				if err := self.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					self.Stop()
					return
				}
			}
		}
	}()

	self.Conn.SetCloseHandler(func(code int, msg string) error {
		log.Warningf("%v: closed %d: %v", self.Session, code, msg)
		return self.Stop()
	})
}

func (self *clientSession) RunCommandChannel() {
	var rw sync.Mutex

	self.prep()
	defer self.Stop()

	log.Warningf("[cmd] Session %v: started command channel", self.Session)

	if id, err := self.Tab.RegisterEventHandler(`*`, func(event *browser.Event) {
		// special case some events so we don't pointlessly send data to the client(s) twice
		switch event.Name {
		case `Page.screencastFrame`:
			return
		}

		rw.Lock()
		defer rw.Unlock()

		if err := self.Conn.WriteJSON(map[string]interface{}{
			`event`:  event.Name,
			`params`: event.P().Value(),
		}); err != nil {
			self.Stop()
		}
	}); err == nil {
		self.waiterId = id
	} else {
		log.Errorf("[%v] Failed to register handler: %v", self.Session, err)
	}

	for {
		if _, msg, err := self.Conn.ReadMessage(); err == nil {
			snippet := string(msg)

			rw.Lock()

			if scope, err := self.Server.env.EvaluateString(snippet); err == nil {
				if err := self.Conn.WriteJSON(map[string]interface{}{
					`success`: true,
					`scope`:   scope.Data(),
				}); err != nil {
					rw.Unlock()
					return
				}
			} else if err := self.Conn.WriteJSON(map[string]interface{}{
				`success`: false,
				`error`:   err.Error(),
			}); err != nil {
				rw.Unlock()
				return
			}

			rw.Unlock()
		} else {
			return
		}
	}
}

func (self *clientSession) Stop() error {
	log.Infof("Removing session %v", self.Session)
	defer self.Tab.RemoveWaiter(self.waiterId)
	defer self.Server.sessions.Delete(self.Session)
	return self.Conn.Close()
}

type Server struct {
	env      *Environment
	server   *negroni.Negroni
	upgrader websocket.Upgrader
	sessions sync.Map
}

func NewServer(env *Environment) *Server {
	return &Server{
		env: env,
	}
}

func (self *Server) ListenAndServe(address string) error {
	if err := self.setupServer(address); err != nil {
		return err
	}

	go func() {
		for {
			self.sessions.Range(func(sid interface{}, v interface{}) bool {
				// c += 1

				if session, ok := v.(*clientSession); ok {
					switch session.Type {
					case ImageSession:
						// skip this round if the client doesn't want the frame yet
						if !session.LastFrameTime.IsZero() && time.Since(session.LastFrameTime) < session.InterframeDelay {
							return true
						}

						if fid, data, width, height := session.Tab.GetMostRecentFrame(); len(data) > 0 {
							if width != session.lastFrameW || height != session.lastFrameH {
								session.lastFrameW = width
								session.lastFrameH = height

								if err := session.Conn.WriteJSON(map[string]interface{}{
									`width`:  width,
									`height`: height,
								}); err != nil {
									session.Stop()
								}
							}

							// don't send duplicate frames
							if fid > session.lastFrameId {
								if err := session.Conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
									session.Stop()
								}

								session.lastFrameId = fid
								session.LastFrameTime = time.Now()
							}

						}
					}
				}

				return true
			})

			time.Sleep(GlobalSessionPollInterval)
		}
	}()

	self.server.Run(address)
	return nil
}

func (self *Server) setupServer(address string) error {
	self.server = negroni.New()
	router := vestigo.NewRouter()

	// setup panic recovery handler
	self.server.Use(negroni.NewRecovery())

	mux := http.NewServeMux()

	ui := diecast.NewServer(`/`, `*.html`)
	ui.BindingPrefix = fmt.Sprintf("http://%s", address)

	if uidir := os.Getenv(`UI`); uidir == `` {
		ui.SetFileSystem(FS(false))
	} else {
		ui.RootPath = uidir
		log.Debugf("[ui] Static asset path is: %v", ui.RootPath)
	}

	if err := ui.Initialize(); err != nil {
		return err
	}

	self.setupRoutes(router)

	mux.Handle(`/api/`, router)
	mux.Handle(`/`, ui)

	self.server.UseHandler(mux)

	return nil
}

func (self *Server) setupRoutes(router *vestigo.Router) {
	router.SetGlobalCors(&vestigo.CorsAccessControl{
		AllowOrigin:      []string{`*`},
		AllowCredentials: true,
		AllowMethods:     []string{`GET`, `POST`, `PUT`, `DELETE`},
		MaxAge:           3600 * time.Second,
		AllowHeaders:     []string{`*`},
	})

	router.Get(`/api/status`, func(w http.ResponseWriter, req *http.Request) {
		httputil.RespondJSON(w, map[string]interface{}{
			`ok`: true,
		})
	})

	router.Post(`/api/tabs/current/script`, func(w http.ResponseWriter, req *http.Request) {
		if scope, err := self.env.EvaluateReader(req.Body); err == nil {
			httputil.RespondJSON(w, scope.Data())
		} else {
			httputil.RespondJSON(w, err, http.StatusBadRequest)
		}
	})

	router.Get(`/api/tabs/current/script`, func(w http.ResponseWriter, req *http.Request) {
		var reqerr error

		if self.env.browser != nil {
			if tab := self.env.browser.Tab(); tab != nil {
				if sid := req.Header.Get(`Sec-Websocket-Protocol`); sid != `` {
					if conn, err := self.upgrader.Upgrade(w, req, http.Header{
						`Sec-Websocket-Protocol`: []string{sid},
					}); err == nil {
						if _, ok := self.sessions.Load(sid); !ok {
							session := &clientSession{
								Type:    CommandSession,
								Tab:     tab,
								Session: sid,
								Conn:    (*websocket.Conn)(conn),
								Server:  self,
							}

							go session.RunCommandChannel()
							self.sessions.Store(sid, session)
						}

						return
					} else {
						reqerr = err
					}
				} else {
					reqerr = fmt.Errorf("Must specify a connection identifier via Sec-Websocket-Protocol")
				}
			} else {
				reqerr = fmt.Errorf("Tab %v does not exist", tab)
			}
		} else {
			reqerr = fmt.Errorf("No browser session available")
		}

		httputil.RespondJSON(w, reqerr)
	})

	router.Get(`/api/tabs/current/screencast`, func(w http.ResponseWriter, req *http.Request) {
		var reqerr error

		if self.env.browser != nil {
			if tab := self.env.browser.Tab(); tab != nil {
				if !tab.IsScreencasting() {
					if err := tab.StartScreencast(
						int(httputil.QInt(req, `q`, 65)),
						int(httputil.QInt(req, `w`, 0)),
						int(httputil.QInt(req, `h`, 0)),
					); err != nil {
						httputil.RespondJSON(w, reqerr, http.StatusConflict)
						return
					}
				}

				if sid := req.Header.Get(`Sec-Websocket-Protocol`); sid != `` {
					if conn, err := self.upgrader.Upgrade(w, req, http.Header{
						`Sec-Websocket-Protocol`: []string{sid},
					}); err == nil {
						if _, ok := self.sessions.Load(sid); !ok {
							session := &clientSession{
								Type:    ImageSession,
								Tab:     tab,
								Session: sid,
								Conn:    (*websocket.Conn)(conn),
								Server:  self,
							}

							if fps := httputil.QInt(req, `fps`); fps > 0 {
								session.InterframeDelay = time.Duration(1000.0/fps) * time.Millisecond
							}

							self.sessions.Store(sid, session)
						}

						return
					} else {
						reqerr = err
					}
				} else {
					reqerr = fmt.Errorf("Must specify a connection identifier via Sec-Websocket-Protocol")
				}
			} else {
				reqerr = fmt.Errorf("Tab %v does not exist", tab)
			}
		} else {
			reqerr = fmt.Errorf("No browser session available")
		}

		httputil.RespondJSON(w, reqerr)
	})

	router.Get(`/api/tabs/current/info`, func(w http.ResponseWriter, req *http.Request) {
		var reqerr error

		if self.env.browser != nil {
			if tab := self.env.browser.Tab(); tab != nil {
				httputil.RespondJSON(w, tab.Info())
			} else {
				reqerr = fmt.Errorf("Tab %v does not exist", tab)
			}
		} else {
			reqerr = fmt.Errorf("No browser session available")
		}

		httputil.RespondJSON(w, reqerr)
	})

	router.Delete(`/api/screencasts/:id`, func(w http.ResponseWriter, req *http.Request) {
		if id := vestigo.Param(req, `id`); id != `` {
			if sessionI, ok := self.sessions.Load(id); ok {
				session := sessionI.(*clientSession)

				httputil.RespondJSON(w, session.Stop())
			} else {
				httputil.RespondJSON(w, fmt.Errorf("Screencast session %v does not exist", id), http.StatusNotFound)
			}
		} else {
			httputil.RespondJSON(w, fmt.Errorf("Screencast session ID not provided"), http.StatusBadRequest)
		}
	})
}
