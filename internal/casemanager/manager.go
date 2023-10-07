package casemanager

import (
	"context"
	"sync"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
)

type Manager struct {
	mutex sync.RWMutex
	cases map[string]map[string]types.Case // methodName -> caseName -> case
}

func NewManager() *Manager {
	return &Manager{
		cases: make(map[string]map[string]types.Case),
	}
}

func (m *Manager) CaseAdd(ctx context.Context, _case types.Case) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.cases[_case.MethodName]; !ok {
		m.cases[_case.MethodName] = make(map[string]types.Case)
	}

	m.cases[_case.MethodName][_case.Name] = _case
	return nil
}

func (m *Manager) CaseDel(ctx context.Context, methodName, name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.cases[methodName]; !ok {
		return nil
	}

	if _, ok := m.cases[methodName][name]; !ok {
		return nil
	}

	delete(m.cases[methodName], name)
	return nil
}

func (m *Manager) CaseGet(ctx context.Context, methodName, name string) (types.Case, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, ok := m.cases[methodName]; !ok {
		return types.Case{}, nil
	}

	if _, ok := m.cases[methodName][name]; !ok {
		return types.Case{}, nil
	}

	return m.cases[methodName][name], nil
}

func (m *Manager) CaseList(ctx context.Context, methodName string) ([]types.Case, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, ok := m.cases[methodName]; !ok {
		return nil, nil
	}

	var cases []types.Case
	for _, _case := range m.cases[methodName] {
		cases = append(cases, _case)
	}

	return cases, nil
}
