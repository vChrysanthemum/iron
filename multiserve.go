package iron

import "log"

type IServer interface {
	ServerName() string
	Serve() error
	Close() error
}

type ServerAsyncCallResult struct {
	Name string
	Err  error
}

type ServerDriver struct {
	servers []IServer
}

func (p *ServerDriver) Init(servers ...IServer) error {
	for i, _ := range servers {
		p.AddServer(servers[i])
	}
	return nil
}

func (p *ServerDriver) AddServer(server IServer) {
	p.servers = append(p.servers, server)
}

func (p *ServerDriver) Serve() error {
	var serverCount = len(p.servers)
	var retChan = make(chan ServerAsyncCallResult, serverCount)
	var err error

	for i := 0; i < serverCount; i++ {
		go func(retChan chan<- ServerAsyncCallResult, server IServer) {
			var ret ServerAsyncCallResult
			ret.Name = server.ServerName()
			ret.Err = server.Serve()
			retChan <- ret
		}(retChan, p.servers[i])
	}

	for i := 0; i < serverCount; i++ {
		var ret = <-retChan
		if ret.Err != nil {
			log.Println("Server serve error, ServerName:", ret.Name, ", Err:", ret.Err)
			err = ret.Err
		}
	}

	return err
}

func (p *ServerDriver) Close() error {
	var serverCount = len(p.servers)
	var retChan = make(chan ServerAsyncCallResult, serverCount)
	var err error

	for i := 0; i < serverCount; i++ {
		go func(retChan chan<- ServerAsyncCallResult, server IServer) {
			var ret ServerAsyncCallResult
			ret.Name = server.ServerName()
			ret.Err = server.Close()
			retChan <- ret
		}(retChan, p.servers[i])
	}

	for i := 0; i < serverCount; i++ {
		var ret = <-retChan
		if ret.Err != nil {
			log.Println("Server close error, ServerName:", ret.Name, ", Err:", ret.Err)
			err = ret.Err
		}
	}

	return err
}
