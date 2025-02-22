package compute

import (
	"context"
	"reflect"

	"github.com/rs/zerolog/log"

	"github.com/bacalhau-project/bacalhau/pkg/bacerrors"
	"github.com/bacalhau-project/bacalhau/pkg/compute/store"
	"github.com/bacalhau-project/bacalhau/pkg/lib/envelope"
	"github.com/bacalhau-project/bacalhau/pkg/lib/ncl"
	"github.com/bacalhau-project/bacalhau/pkg/models"
	"github.com/bacalhau-project/bacalhau/pkg/models/messages"
)

type MessageHandler struct {
	executionStore store.ExecutionStore
}

func NewMessageHandler(executionStore store.ExecutionStore) *MessageHandler {
	return &MessageHandler{
		executionStore: executionStore,
	}
}

func (m *MessageHandler) ShouldProcess(ctx context.Context, message *envelope.Message) bool {
	return message.Metadata.Get(envelope.KeyMessageType) == messages.AskForBidMessageType ||
		message.Metadata.Get(envelope.KeyMessageType) == messages.BidAcceptedMessageType ||
		message.Metadata.Get(envelope.KeyMessageType) == messages.BidRejectedMessageType ||
		message.Metadata.Get(envelope.KeyMessageType) == messages.CancelExecutionMessageType
}

// HandleMessage handles incoming messages
func (m *MessageHandler) HandleMessage(ctx context.Context, message *envelope.Message) error {
	var err error

	switch message.Metadata.Get(envelope.KeyMessageType) {
	case messages.AskForBidMessageType:
		err = m.handleAskForBid(ctx, message)
	case messages.BidAcceptedMessageType:
		err = m.handleBidAccepted(ctx, message)
	case messages.BidRejectedMessageType:
		err = m.handleBidRejected(ctx, message)
	case messages.CancelExecutionMessageType:
		err = m.handleCancel(ctx, message)
	}

	return m.handleError(ctx, message, err)
}

// handleError logs the error with context and returns nil.
// In the future, this can be extended to handle different error types differently.
func (m *MessageHandler) handleError(ctx context.Context, message *envelope.Message, err error) error {
	if err == nil {
		return nil
	}

	// For now, just log the error and return nil
	logger := log.Ctx(ctx).Error()
	for key, value := range message.Metadata.ToMap() {
		logger = logger.Str(key, value)
	}
	logger.Err(err).Msg("Error handling message")
	return nil
}

func (m *MessageHandler) handleAskForBid(ctx context.Context, message *envelope.Message) error {
	request, ok := message.Payload.(*messages.AskForBidRequest)
	if !ok {
		return envelope.NewErrUnexpectedPayloadType("AskForBidRequest", reflect.TypeOf(message.Payload).String())
	}

	// Set the protocol version in the job meta
	execution := request.Execution
	if execution.Job == nil {
		return bacerrors.New("job is missing in the execution").WithComponent(messageHandlerErrorComponent)
	}
	execution.Job.Meta[models.MetaOrchestratorProtocol] = models.ProtocolNCLV1.String()

	// Create the execution
	return m.executionStore.CreateExecution(ctx, *request.Execution)
}

func (m *MessageHandler) handleBidAccepted(ctx context.Context, message *envelope.Message) error {
	request, ok := message.Payload.(*messages.BidAcceptedRequest)
	if !ok {
		return envelope.NewErrUnexpectedPayloadType("BidAcceptedRequest", reflect.TypeOf(message.Payload).String())
	}

	log.Ctx(ctx).Debug().Msgf("bid accepted %s", request.ExecutionID)
	return m.executionStore.UpdateExecutionState(ctx, store.UpdateExecutionRequest{
		ExecutionID: request.ExecutionID,
		Condition: store.UpdateExecutionCondition{
			ExpectedStates: []models.ExecutionStateType{
				models.ExecutionStateNew, models.ExecutionStateAskForBidAccepted},
		},
		NewValues: models.Execution{
			ComputeState: models.NewExecutionState(models.ExecutionStateBidAccepted),
		},
	})
}

func (m *MessageHandler) handleBidRejected(ctx context.Context, message *envelope.Message) error {
	request, ok := message.Payload.(*messages.BidRejectedRequest)
	if !ok {
		return envelope.NewErrUnexpectedPayloadType("BidRejectedRequest", reflect.TypeOf(message.Payload).String())
	}

	log.Ctx(ctx).Debug().Msgf("bid rejected for %s due to %s", request.ExecutionID, request.Message())
	return m.executionStore.UpdateExecutionState(ctx, store.UpdateExecutionRequest{
		ExecutionID: request.ExecutionID,
		NewValues: models.Execution{
			ComputeState: models.NewExecutionState(models.ExecutionStateBidRejected).WithMessage(request.Message()),
		},
		Condition: store.UpdateExecutionCondition{
			ExpectedStates: []models.ExecutionStateType{
				models.ExecutionStateNew, models.ExecutionStateAskForBidAccepted},
		},
		Events: request.Events,
	})
}

func (m *MessageHandler) handleCancel(ctx context.Context, message *envelope.Message) error {
	request, ok := message.Payload.(*messages.CancelExecutionRequest)
	if !ok {
		return envelope.NewErrUnexpectedPayloadType("CancelExecutionRequest", reflect.TypeOf(message.Payload).String())
	}

	log.Ctx(ctx).Debug().Msgf("canceling execution %s due to %s", request.ExecutionID, request.Message())
	return m.executionStore.UpdateExecutionState(ctx, store.UpdateExecutionRequest{
		ExecutionID: request.ExecutionID,
		NewValues: models.Execution{
			ComputeState: models.NewExecutionState(models.ExecutionStateCancelled).WithMessage(request.Message()),
		},
		Events: request.Events,
	})
}

// compile time check for the interface
var _ ncl.MessageHandler = &MessageHandler{}
