package dialmanager

import (
	"context"
	"errors"
	"sync"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/zeromicro/grpc-mock/internal/dialmanager/parser"
)

type Manager struct {
	mutex sync.RWMutex

	upstreams    map[string]*RpcClient
	methodClient map[string]*RpcClient
}

func NewManager() *Manager {
	return &Manager{
		upstreams:    make(map[string]*RpcClient),
		methodClient: make(map[string]*RpcClient),
	}
}

func (m *Manager) AddUpstream(ctx context.Context, upstreams []RpcClientConf) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, upstream := range upstreams {
		cli, err := zrpc.NewClient(upstream.RpcClientConf)
		if err != nil {
			logc.Errorw(ctx, "AddUpstream error", logc.Field("err", err.Error()))
			return err
		}

		desc, err := parser.Parser(cli.Conn())
		if err != nil {

			return err
		}

		client := &RpcClient{
			RpcClientConf: upstream,
			Client:        cli,
			ServicesDesc:  desc,
		}

		for _, svc := range desc {
			for _, method := range svc.Methods {
				m.methodClient[method.FullName] = client
			}
		}

		m.upstreams[upstream.Name] = client
	}

	return nil
}

func (m *Manager) DelUpstream(ctx context.Context, name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, ok := m.upstreams[name]
	if !ok {
		return ErrNotFound
	}

	delete(m.upstreams, name)

	return nil
}

func (m *Manager) Upstream(ctx context.Context, name string) (RpcClientConf, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cli, ok := m.upstreams[name]
	if !ok {
		return RpcClientConf{}, errors.New("not found this upstream")
	}

	return cli.RpcClientConf, nil
}

func (m *Manager) Upstreams(ctx context.Context) ([]RpcClientConf, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var ret []RpcClientConf
	for _, cli := range m.upstreams {
		ret = append(ret, cli.RpcClientConf)
	}

	return ret, nil
}

func (m *Manager) Methods(ctx context.Context) (map[string][]string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var resp = map[string][]string{}

	for _, client := range m.upstreams {
		for _, svc := range client.ServicesDesc {
			for _, method := range svc.Methods {
				resp[svc.FullName] = append(resp[svc.FullName], method.FullName)
			}
		}
	}

	return resp, nil
}

func (m *Manager) MethodDetail(ctx context.Context, method string) (parser.MethodDesc, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cli, ok := m.methodClient[method]
	if !ok {
		return parser.MethodDesc{}, ErrNotFound
	}

	for _, svc := range cli.ServicesDesc {
		for _, m := range svc.Methods {
			if m.FullName == method {
				return m, nil
			}
		}
	}

	return parser.MethodDesc{}, ErrNotFound
}

func (m *Manager) UpstreamClient(ctx context.Context, name string) (grpc.ClientConnInterface, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cli, ok := m.methodClient[name]
	if !ok {
		return nil, ErrNotFound
	}

	return cli.Conn(), nil
}
