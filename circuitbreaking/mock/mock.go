package mock

import "github.com/stretchr/testify/mock"

type MockCircuitBreaker struct {
	mock.Mock
}

func (n *MockCircuitBreaker) Failed() {
	n.Called()
}

func (n *MockCircuitBreaker) Succeeded() {
	n.Called()
}

func (n *MockCircuitBreaker) CanProceed() bool {
	return n.Called().Bool(0)
}

func (n *MockCircuitBreaker) CannotProceed() bool {
	return n.Called().Bool(0)
}
