package stun

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"

	"github.com/agan83/netdoop/utils"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gortc.io/stun"
)

var errNotSTUNMessage = errors.New("not stun message")
var software = stun.NewSoftware("stund")

type STUNServer interface {
	Run()
	Close()
}

type stunServer struct {
	logger       *zap.Logger
	env          *viper.Viper
	conn         net.PacketConn
	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

func NewSTUNServer() STUNServer {
	s := &stunServer{
		shutdownCh: make(chan struct{}),
	}
	s.logger = utils.GetLogger().Named("stun")
	s.env = utils.GetEnv()

	network := s.env.GetString("stun_network")
	addr := s.env.GetString("stun_addr")
	conn, err := net.ListenPacket(network, addr)
	if err != nil {
		s.logger.Error("listen packet", zap.Error(err))
		return nil
	}
	s.conn = conn
	return s
}

func (s *stunServer) Close() {
	s.shutdownLock.Lock()
	defer s.shutdownLock.Unlock()

	if s.shutdown {
		return
	}
	s.logger.Debug("server close")
	s.shutdown = true
	close(s.shutdownCh)
}

func (s *stunServer) Run() {
	s.logger.Debug("server running")
	defer s.logger.Debug("server stopped")

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var (
			res = new(stun.Message)
			req = new(stun.Message)
		)
		for {
			if s.shutdown {
				return
			}
			if err := s.serveConn(res, req); err != nil {
				s.logger.Error("serve conn", zap.Error(err))
				continue
			}
			res.Reset()
			req.Reset()
		}
	}()

	<-s.shutdownCh

	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *stunServer) serveConn(res, req *stun.Message) error {
	if s.conn == nil {
		return nil
	}
	buf := make([]byte, 1024)
	n, addr, err := s.conn.ReadFrom(buf)
	if err != nil {
		return errors.Wrap(err, "ReadFrom")
	}
	needChange := false
	if binary.BigEndian.Uint32(buf[4:8]) == 0x30313031 {
		needChange = true
		binary.BigEndian.PutUint32(buf[4:8], 0x2112A442)
	}

	if _, err := req.Write(buf[:n]); err != nil {
		return errors.Wrap(err, "Write")
	}
	if err := s.basicProcess(addr, buf[:n], req, res); err != nil {
		if err == errNotSTUNMessage {
			return nil
		}
		return errors.Wrap(err, "basicProcess")
	}
	if needChange {
		binary.BigEndian.PutUint32(res.Raw[4:8], 0x30313031)
	}
	if _, err := s.conn.WriteTo(res.Raw, addr); err != nil {
		return errors.Wrap(err, "WriteTo")
	}
	s.logger.Warn("debug", zap.Any("req", req), zap.Any("resp", res))

	return nil
}

func (s *stunServer) basicProcess(addr net.Addr, b []byte, req, res *stun.Message) error {
	if !stun.IsMessage(b) {
		return errNotSTUNMessage
	}
	if _, err := req.Write(b); err != nil {
		return errors.Wrap(err, "failed to read message")
	}
	var (
		ip   net.IP
		port int
	)
	switch a := addr.(type) {
	case *net.UDPAddr:
		ip = a.IP
		port = a.Port
	default:
		return errors.New(fmt.Sprintf("unknown addr: %v", addr))
	}
	s.logger.Debug("debug", zap.String("ip", ip.String()), zap.Int("port", port))
	return res.Build(req,
		stun.BindingSuccess,
		software,
		&stun.XORMappedAddress{
			IP:   ip,
			Port: port,
		},
		stun.Fingerprint,
	)
}
