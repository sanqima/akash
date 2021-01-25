package handler

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"

	sdk "github.com/cosmos/cosmos-sdk/types"

	atypes "github.com/ovrclk/akash/x/audit/types"
	etypes "github.com/ovrclk/akash/x/escrow/types"
	"github.com/ovrclk/akash/x/market/types"
	ptypes "github.com/ovrclk/akash/x/provider/types"
)

const (
	bidEscrowScope = "bid"
)

type msgServer struct {
	keepers Keepers
}

// NewMsgServerImpl returns an implementation of the market MsgServer interface
// for the provided Keeper.
func NewServer(k Keepers) types.MsgServer {
	return &msgServer{keepers: k}
}

var _ types.MsgServer = msgServer{}

func (ms msgServer) CreateBid(goCtx context.Context, msg *types.MsgCreateBid) (*types.MsgCreateBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	order, found := ms.keepers.Market.GetOrder(ctx, msg.Order)
	if !found {
		return nil, types.ErrInvalidOrder
	}

	if err := order.ValidateCanBid(); err != nil {
		return nil, err
	}

	if !msg.Price.IsValid() {
		return nil, types.ErrBidInvalidPrice
	}

	if order.Price().IsLT(msg.Price) {
		return nil, types.ErrBidOverOrder
	}

	provider, err := sdk.AccAddressFromBech32(msg.Provider)
	if err != nil {
		return nil, types.ErrEmptyProvider
	}

	var prov ptypes.Provider
	if prov, found = ms.keepers.Provider.Get(ctx, provider); !found {
		return nil, types.ErrEmptyProvider
	}

	provAttr, _ := ms.keepers.Audit.GetProviderAttributes(ctx, provider)

	provAttr = append([]atypes.Provider{{
		Owner:      msg.Provider,
		Attributes: prov.Attributes,
	}}, provAttr...)

	if !order.MatchRequirements(provAttr) {
		return nil, types.ErrAttributeMismatch
	}

	bid, err := ms.keepers.Market.CreateBid(ctx, msg.Order, provider, msg.Price)
	if err != nil {
		return nil, err
	}

	// crate escrow account for this bid
	// todo: check deposit
	if err := ms.keepers.Escrow.AccountCreate(ctx, etypes.AccountID{
		Scope: bidEscrowScope,
		XID:   bid.ID().String(),
	}, bid.ID().Owner, sdk.NewCoin("XXX", sdk.NewInt(0))); err != nil {
		return &types.MsgCreateBidResponse{}, err
	}

	telemetry.IncrCounter(1.0, "akash.bids")
	return &types.MsgCreateBidResponse{}, nil
}

func (ms msgServer) CloseBid(goCtx context.Context, msg *types.MsgCloseBid) (*types.MsgCloseBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	bid, found := ms.keepers.Market.GetBid(ctx, msg.BidID)
	if !found {
		return nil, types.ErrUnknownBid
	}

	order, found := ms.keepers.Market.GetOrder(ctx, msg.BidID.OrderID())
	if !found {
		return nil, types.ErrUnknownOrderForBid
	}

	if bid.State == types.BidOpen {
		ms.keepers.Market.OnBidClosed(ctx, bid)
		return &types.MsgCloseBidResponse{}, nil
	}

	lease, found := ms.keepers.Market.GetLease(ctx, types.LeaseID(msg.BidID))
	if !found {
		return nil, types.ErrUnknownLeaseForBid
	}

	if lease.State != types.LeaseActive {
		return nil, types.ErrLeaseNotActive
	}

	if bid.State != types.BidActive {
		return nil, types.ErrBidNotActive
	}

	ms.keepers.Market.OnBidClosed(ctx, bid)
	ms.keepers.Market.OnLeaseClosed(ctx, lease)
	ms.keepers.Market.OnOrderClosed(ctx, order)
	ms.keepers.Deployment.OnBidClosed(ctx, order.ID().GroupID())
	telemetry.IncrCounter(1.0, "akash.order_closed")

	return &types.MsgCloseBidResponse{}, nil
}

func (ms msgServer) CreateLease(goCtx context.Context, msg *types.MsgCreateLease) (*types.MsgCreateLeaseResponse, error) {
	// TODO
	return &types.MsgCreateLeaseResponse{}, nil
}

func (ms msgServer) CloseOrder(goCtx context.Context, msg *types.MsgCloseOrder) (*types.MsgCloseOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// close payment

	order, found := ms.keepers.Market.GetOrder(ctx, msg.OrderID)
	if !found {
		return nil, types.ErrUnknownOrder
	}

	lease, found := ms.keepers.Market.LeaseForOrder(ctx, order.ID())
	if !found {
		return nil, types.ErrNoLeaseForOrder
	}

	ms.keepers.Market.OnOrderClosed(ctx, order)
	ms.keepers.Market.OnLeaseClosed(ctx, lease)
	ms.keepers.Deployment.OnOrderClosed(ctx, order.ID().GroupID())

	return &types.MsgCloseOrderResponse{}, nil
}
