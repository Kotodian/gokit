package server

import (
	"fmt"
	"net"
	"time"

	"github.com/Kotodian/gokit/datasource/grpc/interceptor"
	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

var (
	logEntry *log.Entry
	server   *grpc.Server
)

func init() {
	server = grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.RPCServerInterceptor),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     5 * time.Minute,  // <--- This fixes it!
			MaxConnectionAge:      10 * time.Minute, // <--- This fixes it!
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
			PermitWithoutStream: true,            // Allow pings even when there are no active streams
		}))
}

func GetServer() *grpc.Server {
	return server
}

type IServer interface {
	Init()
	// Run()
	// Stop()
	RegisterService(fn func(s *grpc.Server, srv interface{}))
}

type Base struct {
}

func (b *Base) RegisterService(fn func(s *grpc.Server, srv interface{})) {
	fn(server, b)
}

func (b *Base) Init() {
}

func Run(s IServer, addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
		// return
	}
	s.Init()
	server.Serve(listener)
}

func Stop() {
	server.GracefulStop()
}

// func (s *Base) RegisterService(fn func(s *grpc.Server, srv interface{})) {
// 	fn(s.s, s)
// }

// func bootInit() error {
// 	server := NewRpcServer(fmt.Sprintf("0.0.0.0:%d", 8070))
// 	go func() {
// 		server.Run()
// 	}()
// 	return nil
// }
