package celeritas

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
)

var maintenanceMode bool

type RPCServer struct {
}

func (r *RPCServer) MaintenanceMode(enable bool, response *string) error {
	if enable {
		maintenanceMode = true
		*response = "Server in maintenance mode"
	} else {
		maintenanceMode = false
		*response = "Server live"
	}
	return nil
}

func (c *Celeritas) listenRPC() {
	if port := os.Getenv("RPC_PORT"); port != "" {
		c.InfoLog.Println("Starting RPC server on port", port)
		if err := rpc.Register(new(RPCServer)); err != nil {
			c.ErrorLog.Println(err)
			return
		}
		listen, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", port))
		if err != nil {
			c.ErrorLog.Println(err)
			return
		}
		for {
			conn, err := listen.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(conn)
		}
	}
}
