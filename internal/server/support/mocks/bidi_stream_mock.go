// Code generated by MockGen. DO NOT EDIT.
// Source: google.golang.org/grpc (interfaces: BidiStreamingServer)
//
// Generated by this command:
//
//	mockgen -destination=./bidi_stream_mock.go -package=mocks google.golang.org/grpc BidiStreamingServer
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	metadata "google.golang.org/grpc/metadata"
)

// MockBidiStreamingServer is a mock of BidiStreamingServer interface.
type MockBidiStreamingServer[Req any, Res any] struct {
	ctrl     *gomock.Controller
	recorder *MockBidiStreamingServerMockRecorder[Req, Res]
	isgomock struct{}
}

// MockBidiStreamingServerMockRecorder is the mock recorder for MockBidiStreamingServer.
type MockBidiStreamingServerMockRecorder[Req any, Res any] struct {
	mock *MockBidiStreamingServer[Req, Res]
}

// NewMockBidiStreamingServer creates a new mock instance.
func NewMockBidiStreamingServer[Req any, Res any](ctrl *gomock.Controller) *MockBidiStreamingServer[Req, Res] {
	mock := &MockBidiStreamingServer[Req, Res]{ctrl: ctrl}
	mock.recorder = &MockBidiStreamingServerMockRecorder[Req, Res]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBidiStreamingServer[Req, Res]) EXPECT() *MockBidiStreamingServerMockRecorder[Req, Res] {
	return m.recorder
}

// Context mocks base method.
func (m *MockBidiStreamingServer[Req, Res]) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context.
func (mr *MockBidiStreamingServerMockRecorder[Req, Res]) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockBidiStreamingServer[Req, Res])(nil).Context))
}

// Recv mocks base method.
func (m *MockBidiStreamingServer[Req, Res]) Recv() (*Req, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recv")
	ret0, _ := ret[0].(*Req)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Recv indicates an expected call of Recv.
func (mr *MockBidiStreamingServerMockRecorder[Req, Res]) Recv() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recv", reflect.TypeOf((*MockBidiStreamingServer[Req, Res])(nil).Recv))
}

// RecvMsg mocks base method.
func (m_2 *MockBidiStreamingServer[Req, Res]) RecvMsg(m any) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "RecvMsg", m)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg.
func (mr *MockBidiStreamingServerMockRecorder[Req, Res]) RecvMsg(m any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockBidiStreamingServer[Req, Res])(nil).RecvMsg), m)
}

// Send mocks base method.
func (m *MockBidiStreamingServer[Req, Res]) Send(arg0 *Res) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockBidiStreamingServerMockRecorder[Req, Res]) Send(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBidiStreamingServer[Req, Res])(nil).Send), arg0)
}

// SendHeader mocks base method.
func (m *MockBidiStreamingServer[Req, Res]) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader.
func (mr *MockBidiStreamingServerMockRecorder[Req, Res]) SendHeader(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockBidiStreamingServer[Req, Res])(nil).SendHeader), arg0)
}

// SendMsg mocks base method.
func (m_2 *MockBidiStreamingServer[Req, Res]) SendMsg(m any) error {
	m_2.ctrl.T.Helper()
	ret := m_2.ctrl.Call(m_2, "SendMsg", m)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg.
func (mr *MockBidiStreamingServerMockRecorder[Req, Res]) SendMsg(m any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockBidiStreamingServer[Req, Res])(nil).SendMsg), m)
}

// SetHeader mocks base method.
func (m *MockBidiStreamingServer[Req, Res]) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader.
func (mr *MockBidiStreamingServerMockRecorder[Req, Res]) SetHeader(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockBidiStreamingServer[Req, Res])(nil).SetHeader), arg0)
}

// SetTrailer mocks base method.
func (m *MockBidiStreamingServer[Req, Res]) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer.
func (mr *MockBidiStreamingServerMockRecorder[Req, Res]) SetTrailer(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockBidiStreamingServer[Req, Res])(nil).SetTrailer), arg0)
}
