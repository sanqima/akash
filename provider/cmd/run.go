package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/ovrclk/akash/provider/bidengine"
	"os"
	"time"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/go-kit/kit/log/term"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	"golang.org/x/sync/errgroup"

	"github.com/ovrclk/akash/client"
	"github.com/ovrclk/akash/cmd/common"
	"github.com/ovrclk/akash/events"
	"github.com/ovrclk/akash/provider"
	"github.com/ovrclk/akash/provider/cluster"
	"github.com/ovrclk/akash/provider/cluster/kube"
	"github.com/ovrclk/akash/provider/gateway"
	"github.com/ovrclk/akash/provider/session"
	"github.com/ovrclk/akash/pubsub"
	dmodule "github.com/ovrclk/akash/x/deployment"
	mmodule "github.com/ovrclk/akash/x/market"
	pmodule "github.com/ovrclk/akash/x/provider"
	ptypes "github.com/ovrclk/akash/x/provider/types"
)

const (
	// FlagClusterK8s informs the provider to scan and utilize localized kubernetes client configuration
	FlagClusterK8s = "cluster-k8s"
	// FlagK8sManifestNS
	FlagK8sManifestNS = "k8s-manifest-ns"
	// FlagGatewayListenAddress determines listening address for Manifests
	FlagGatewayListenAddress       = "gateway-listen-address"
	FlagBidPricingStrategy         = "bid-price-strategy"
	FlagBidPriceCPUScale           = "bid-price-cpu-scale"
	FlagBidPriceMemoryScale        = "bid-price-memory-scale"
	FlagBidPriceStorageScale       = "bid-price-storage-scale"
	FlagBidPriceScriptPath         = "bid-price-script-path"
	FlagBidPriceScriptProcessLimit = "bid-price-script-process-limit"
	FlagBidPriceScriptTimeout      = "bid-price-script-process-timeout"
)

var (
	errInvalidConfig = errors.New("Invalid configuration")
)

// RunCmd launches the Akash Provider service
func RunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run akash provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.RunForever(func(ctx context.Context) error {
				return doRunCmd(ctx, cmd, args)
			})
		},
	}

	cmd.Flags().String(flags.FlagChainID, "", "The network chain ID")
	if err := viper.BindPFlag(flags.FlagChainID, cmd.Flags().Lookup(flags.FlagChainID)); err != nil {
		return nil
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().Bool(FlagClusterK8s, false, "Use Kubernetes cluster")
	if err := viper.BindPFlag(FlagClusterK8s, cmd.Flags().Lookup(FlagClusterK8s)); err != nil {
		return nil
	}

	cmd.Flags().String(FlagK8sManifestNS, "lease", "Cluster manifest namespace")
	if err := viper.BindPFlag(FlagK8sManifestNS, cmd.Flags().Lookup(FlagK8sManifestNS)); err != nil {
		return nil
	}

	cmd.Flags().String(FlagGatewayListenAddress, "0.0.0.0:8080", "Gateway listen address")
	if err := viper.BindPFlag(FlagGatewayListenAddress, cmd.Flags().Lookup(FlagGatewayListenAddress)); err != nil {
		return nil
	}

	cmd.Flags().String(FlagBidPricingStrategy, "scale", "Pricing strategy to use")
	if err := viper.BindPFlag(FlagBidPricingStrategy, cmd.Flags().Lookup(FlagBidPricingStrategy)); err != nil {
		return nil
	}

	cmd.Flags().Uint64(FlagBidPriceCPUScale, 0, "cpu pricing scale in uakt")
	if err := viper.BindPFlag(FlagBidPriceCPUScale, cmd.Flags().Lookup(FlagBidPriceCPUScale)); err != nil {
		return nil
	}
	cmd.Flags().Uint64(FlagBidPriceMemoryScale, 0, "memory pricing scale in uakt")
	if err := viper.BindPFlag(FlagBidPriceMemoryScale, cmd.Flags().Lookup(FlagBidPriceMemoryScale)); err != nil {
		return nil
	}
	cmd.Flags().Uint64(FlagBidPriceStorageScale, 0, "storage pricing scale in uakt")
	if err := viper.BindPFlag(FlagBidPriceStorageScale, cmd.Flags().Lookup(FlagBidPriceStorageScale)); err != nil {
		return nil
	}
	cmd.Flags().String(FlagBidPriceScriptPath, "", "path to script to run for computing bid price")
	if err := viper.BindPFlag(FlagBidPriceScriptPath, cmd.Flags().Lookup(FlagBidPriceScriptPath)); err != nil {
		return nil
	}
	cmd.Flags().Uint(FlagBidPriceScriptProcessLimit, 32, "limit to the number of scripts run concurrently for bid pricing")
	if err := viper.BindPFlag(FlagBidPriceScriptProcessLimit, cmd.Flags().Lookup(FlagBidPriceScriptProcessLimit)); err != nil {
		return nil
	}
	cmd.Flags().String(FlagBidPriceScriptTimeout, "10000ms", "execution timelimit for bid pricing as a duration")
	if err := viper.BindPFlag(FlagBidPriceScriptTimeout, cmd.Flags().Lookup(FlagBidPriceScriptTimeout)); err != nil {
		return nil
	}

	return cmd
}

const (
	bidPricingStrategyScale       = "scale"
	bidPricingStrategyRandomRange = "randomRange"
	bidPricingStrategyShellScript = "shellScript"
)

var allowedBidPricingStrategies = [...]string{
	bidPricingStrategyScale,
	bidPricingStrategyRandomRange,
	bidPricingStrategyShellScript,
}

var errNoSuchBidPricingStrategy = fmt.Errorf("No such bid pricing strategy. Allowed: %v", allowedBidPricingStrategies)

func createBidPricingStrategy(cmd *cobra.Command, strategy string) (bidengine.BidPricingStrategy, error) {
	if strategy == bidPricingStrategyScale {
		cpuScale, err := cmd.Flags().GetUint64(FlagBidPriceCPUScale)
		if err != nil {
			return nil, err
		}
		memoryScale, err := cmd.Flags().GetUint64(FlagBidPriceMemoryScale)
		if err != nil {
			return nil, err
		}
		storageScale, err := cmd.Flags().GetUint64(FlagBidPriceStorageScale)
		if err != nil {
			return nil, err
		}
		return bidengine.MakeScalePricing(cpuScale, memoryScale, storageScale)
	}

	if strategy == bidPricingStrategyRandomRange {
		return bidengine.MakeRandomRangePricing()
	}

	if strategy == bidPricingStrategyShellScript {
		scriptPath, err := cmd.Flags().GetString(FlagBidPriceScriptPath)
		if err != nil {
			return nil, err
		}

		processLimit, err := cmd.Flags().GetUint(FlagBidPriceScriptProcessLimit)
		if err != nil {
			return nil, err
		}

		runtimeLimitAsString, err := cmd.Flags().GetString(FlagBidPriceScriptTimeout)
		if err != nil {
			return nil, err
		}

		runtimeLimit, err := time.ParseDuration(runtimeLimitAsString)
		if err != nil {
			return nil, err
		}

		return bidengine.MakeShellScriptPricing(scriptPath, processLimit, runtimeLimit)
	}

	return nil, errNoSuchBidPricingStrategy
}

// doRunCmd initializes all of the Provider functionality, hangs, and awaits shutdown signals.
func doRunCmd(ctx context.Context, cmd *cobra.Command, _ []string) error {

	strategy, err := cmd.Flags().GetString(FlagBidPricingStrategy)
	if err != nil {
		return err
	}

	pricing, err := createBidPricingStrategy(cmd, strategy)
	if err != nil {
		return err
	}

	cctx := sdkclient.GetClientContextFromCmd(cmd)

	from, _ := cmd.Flags().GetString(flags.FlagFrom)
	_, _, err = cosmosclient.GetFromFields(cctx.Keyring, from, false)
	if err != nil {
		return err
	}

	cctx, err = sdkclient.ReadTxCommandFlags(cctx, cmd.Flags())
	if err != nil {
		return err
	}

	txFactory := tx.NewFactoryCLI(cctx, cmd.Flags()).WithTxConfig(cctx.TxConfig).WithAccountRetriever(cctx.AccountRetriever)

	keyname := cctx.GetFromName()
	info, err := txFactory.Keybase().Key(keyname)
	if err != nil {
		return err
	}

	gwaddr := viper.GetString(FlagGatewayListenAddress)

	log := openLogger()

	// TODO: actually get the passphrase?
	// passphrase, err := keys.GetPassphrase(fromName)
	aclient := client.NewClient(
		log,
		cctx,
		txFactory,
		info,
		keys.DefaultKeyPass,
		client.NewQueryClient(
			dmodule.AppModuleBasic{}.GetQueryClient(cctx),
			mmodule.AppModuleBasic{}.GetQueryClient(cctx),
			pmodule.AppModuleBasic{}.GetQueryClient(cctx),
		),
	)

	res, err := aclient.Query().Provider(
		context.Background(),
		&ptypes.QueryProviderRequest{Owner: info.GetAddress().String()},
	)
	if err != nil {
		return err
	}

	pinfo := &res.Provider

	// k8s client creation
	cclient, err := createClusterClient(log, cmd, pinfo.HostURI)
	if err != nil {
		return err
	}

	session := session.New(log, aclient, pinfo)

	if err := cctx.Client.Start(); err != nil {
		return err
	}

	bus := pubsub.NewBus()
	defer bus.Close()

	group, ctx := errgroup.WithContext(ctx)

	service, err := provider.NewService(ctx, session, bus, cclient, pricing)
	if err != nil {
		return group.Wait()
	}

	gateway := gateway.NewServer(ctx, log, service, gwaddr)

	group.Go(func() error {
		return events.Publish(ctx, cctx.Client, "provider-cli", bus)
	})

	group.Go(func() error {
		<-service.Done()
		return nil
	})

	group.Go(gateway.ListenAndServe)

	group.Go(func() error {
		<-ctx.Done()
		return gateway.Close()
	})

	err = group.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func openLogger() log.Logger {
	// logger with no color output - current debug colors are invisible for me.
	return log.NewTMLoggerWithColorFn(log.NewSyncWriter(os.Stdout), func(_ ...interface{}) term.FgBgColor {
		return term.FgBgColor{}
	})
}

func createClusterClient(log log.Logger, _ *cobra.Command, host string) (cluster.Client, error) {
	if !viper.GetBool(FlagClusterK8s) {
		// Condition that there is no Kubernetes API to work with.
		return cluster.NullClient(), nil
	}
	ns := viper.GetString(FlagK8sManifestNS)
	if ns == "" {
		return nil, fmt.Errorf("%w: --%s required", errInvalidConfig, FlagK8sManifestNS)
	}
	return kube.NewClient(log, host, ns)
}
