package mocks

import (
	"github.com/nelsonmarro/verith/internal/sri"
	"github.com/stretchr/testify/mock"
)

type MockSriClient struct {
	mock.Mock
}

func (m *MockSriClient) EnviarComprobante(xmlFirmado []byte, environment int) (*sri.RespuestaRecepcion, error) {
	args := m.Called(xmlFirmado, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sri.RespuestaRecepcion), args.Error(1)
}

func (m *MockSriClient) AutorizarComprobante(claveAcceso string, environment int) (*sri.RespuestaAutorizacion, error) {
	args := m.Called(claveAcceso, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sri.RespuestaAutorizacion), args.Error(1)
}
