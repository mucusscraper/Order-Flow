package domain

import (
	"testing"
)

func TestOrderStateTransitions(t *testing.T) {
	order := &Order{
		Status: StatusCreated,
	}
	if !order.CanTransitionTo(StatusPaymentPending) {
		t.Errorf("expected transition from CREATED to PAYMENT_PENDING to be valid")
	}
	if order.CanTransitionTo(StatusCompleted) {
		t.Errorf("expected transition from CREATED to COMPLETED to be invalid")
	}
}

func TestOrderCancellationRules(t *testing.T) {
	order := &Order{Status: StatusCreated}
	if !order.CanTransitionTo(StatusCancelled) {
		t.Errorf("expected CREATED order to be cancellable")
	}
	completedOrder := &Order{Status: StatusCompleted}
	if completedOrder.CanTransitionTo(StatusCancelled) {
		t.Errorf("expected COMPLETED order NOT to be cancellable")
	}
}
