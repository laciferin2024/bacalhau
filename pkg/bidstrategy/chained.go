package bidstrategy

import (
	"context"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/bacalhau-project/bacalhau/pkg/models"
)

type ChainedBidStrategy struct {
	Semantics []SemanticBidStrategy
	Resources []ResourceBidStrategy
}

type StrategyOpt func(strategy *ChainedBidStrategy)

func WithSemantics(strategy ...SemanticBidStrategy) StrategyOpt {
	return func(chain *ChainedBidStrategy) {
		chain.Semantics = append(chain.Semantics, strategy...)
	}
}

func WithResources(strategy ...ResourceBidStrategy) StrategyOpt {
	return func(chain *ChainedBidStrategy) {
		chain.Resources = append(chain.Resources, strategy...)
	}
}

func NewChainedBidStrategy(opts ...StrategyOpt) *ChainedBidStrategy {
	out := &ChainedBidStrategy{}
	for _, o := range opts {
		o(out)
	}
	return out
}

// AddStrategy Add new strategy to the end of the chain
// NOTE: this is not thread safe.
func (c *ChainedBidStrategy) AddStrategy(opts ...StrategyOpt) {
	for _, o := range opts {
		o(c)
	}
}

// ShouldBid Iterate over all strategies, and return shouldBid if no error is thrown
// and none of the strategies return should not bid.
func (c *ChainedBidStrategy) ShouldBid(ctx context.Context, request BidStrategyRequest) (BidStrategyResponse, error) {
	reasons := make([]string, 0, len(c.Semantics))

	for _, strategy := range c.Semantics {
		response, err := strategy.ShouldBid(ctx, request)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("error asking bidding strategy %s if we should bid",
				reflect.TypeOf(strategy).String())
			return BidStrategyResponse{}, err
		}
		var status string
		if !response.ShouldBid {
			status = "should not bid"
		}
		if status != "" {
			log.Ctx(ctx).Debug().Msgf("bidding strategy %s returned %s due to: %s",
				reflect.TypeOf(strategy).String(), status, response.Reason)
			return response, nil
		}
		reasons = append(reasons, response.Reason)
	}

	return BidStrategyResponse{ShouldBid: true, Reason: strings.Join(reasons, "; ")}, nil
}

// ShouldBidBasedOnUsage Iterate over all strategies, and return shouldBid if no error is thrown
// and none of the strategies return should not bid.
func (c *ChainedBidStrategy) ShouldBidBasedOnUsage(
	ctx context.Context, request BidStrategyRequest, usage models.Resources) (BidStrategyResponse, error) {
	reasons := make([]string, 0, len(c.Resources))
	for _, strategy := range c.Resources {
		response, err := strategy.ShouldBidBasedOnUsage(ctx, request, usage)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("error asking bidding strategy %s if we should bid",
				reflect.TypeOf(strategy).String())
			return BidStrategyResponse{}, err
		}
		var status string
		if !response.ShouldBid {
			status = "should not bid"
		}
		if status != "" {
			log.Ctx(ctx).Debug().Msgf("bidding strategy %s returned %s due to: %s",
				reflect.TypeOf(strategy).String(), status, response.Reason)
			return response, nil
		}
	}

	return BidStrategyResponse{ShouldBid: true, Reason: strings.Join(reasons, "; ")}, nil
}

var _ SemanticBidStrategy = (*ChainedBidStrategy)(nil)
var _ ResourceBidStrategy = (*ChainedBidStrategy)(nil)
