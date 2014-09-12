// Commands from http://redis.io/commands#connection

package miniredis

import (
	"strconv"

	"github.com/bsm/redeo"
)

// Select sets the DB id for all direct commands.
func (m *Miniredis) Select(i int) {
	m.Lock()
	defer m.Unlock()
	m.selectedDB = i
}

func commandsConnection(m *Miniredis, srv *redeo.Server) {
	srv.HandleFunc("AUTH", m.cmdAuth)
	srv.HandleFunc("ECHO", m.cmdEcho)
	srv.HandleFunc("PING", m.cmdPing)
	srv.HandleFunc("SELECT", m.cmdSelect)
}

// PING
func (m *Miniredis) cmdPing(out *redeo.Responder, r *redeo.Request) error {
	out.WriteInlineString("PONG")
	return nil
}

// AUTH
func (m *Miniredis) cmdAuth(out *redeo.Responder, r *redeo.Request) error {
	out.WriteOK()
	return nil
}

// ECHO
func (m *Miniredis) cmdEcho(out *redeo.Responder, r *redeo.Request) error {
	if len(r.Args) != 1 {
		setDirty(r.Client())
		out.WriteErrorString("usage error")
		return nil
	}
	msg := r.Args[0]
	out.WriteString(msg)
	return nil
}

// SELECT
func (m *Miniredis) cmdSelect(out *redeo.Responder, r *redeo.Request) error {
	if len(r.Args) != 1 {
		setDirty(r.Client())
		out.WriteErrorString("usage error")
		return nil
	}
	id, err := strconv.Atoi(r.Args[0])
	if err != nil {
		id = 0
	}

	m.Lock()
	defer m.Unlock()

	ctx := getCtx(r.Client())
	ctx.selectedDB = id

	out.WriteOK()
	return nil
}
