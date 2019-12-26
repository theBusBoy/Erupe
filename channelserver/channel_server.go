package channelserver

import (
	"database/sql"
	"fmt"
	"net"
	"sync"
)

// Config struct allows configuring the server.
type Config struct {
	DB         *sql.DB
	ListenAddr string
}

// Server is a MHF channel server.
type Server struct {
	sync.Mutex
	acceptConns chan net.Conn
	deleteConns chan net.Conn
	sessions    map[net.Conn]*Session
	db          *sql.DB
	listenAddr  string
	listener    net.Listener // Listener that is created when Server.Start is called.
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		acceptConns: make(chan net.Conn),
		deleteConns: make(chan net.Conn),
		sessions:    make(map[net.Conn]*Session),
		db:          config.DB,
		listenAddr:  config.ListenAddr,
	}
	return s
}

// Start starts the server in a new goroutine.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	s.listener = l
	//defer l.Close()

	go s.acceptClients()
	go s.manageSessions()

	return nil
}

func (s *Server) acceptClients() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// TODO(Andoryuuta): Implement shutdown logic to end this goroutine cleanly here.
			fmt.Println(err)
			continue
		}
		s.acceptConns <- conn
	}
}

func (s *Server) manageSessions() {
	for {
		select {
		case newConn := <-s.acceptConns:
			session := NewSession(s, newConn)

			s.Lock()
			s.sessions[newConn] = session
			s.Unlock()

			session.Start()

		case delConn := <-s.deleteConns:
			s.Lock()
			delete(s.sessions, delConn)
			s.Unlock()
		}
	}
}