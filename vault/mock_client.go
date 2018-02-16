// Code generated by MockGen. DO NOT EDIT.
// Source: vault/client.go

// Package mock_vault is a generated GoMock package.
package vault

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockVaultClient is a mock of VaultClient interface
type MockVaultClient struct {
	ctrl     *gomock.Controller
	recorder *MockVaultClientMockRecorder
}

// MockVaultClientMockRecorder is the mock recorder for MockVaultClient
type MockVaultClientMockRecorder struct {
	mock *MockVaultClient
}

// NewMockVaultClient creates a new mock instance
func NewMockVaultClient(ctrl *gomock.Controller) *MockVaultClient {
	mock := &MockVaultClient{ctrl: ctrl}
	mock.recorder = &MockVaultClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockVaultClient) EXPECT() *MockVaultClientMockRecorder {
	return m.recorder
}

// Exists mocks base method
func (m *MockVaultClient) Exists(prefix, team, pipeline, mapKey, keyName string) (bool, error) {
	ret := m.ctrl.Call(m, "Exists", prefix, team, pipeline, mapKey, keyName)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exists indicates an expected call of Exists
func (mr *MockVaultClientMockRecorder) Exists(prefix, team, pipeline, mapKey, keyName interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockVaultClient)(nil).Exists), prefix, team, pipeline, mapKey, keyName)
}
