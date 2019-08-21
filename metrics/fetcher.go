package metrics

import (
	"context"
	"fmt"

	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

//go:generate counterfeiter . LogCacheClient

type LogCacheClient interface {
	PromQL(ctx context.Context, query string, opts ...logcache.PromQLOption) (*logcache_v1.PromQL_InstantQueryResult, error)
}

type LogCacheFetcher struct {
	logCacheClient LogCacheClient
}

func NewLogCacheFetcher(logCacheClient LogCacheClient) LogCacheFetcher {
	return LogCacheFetcher{logCacheClient: logCacheClient}
}

func (f LogCacheFetcher) FetchInstanceEntitlementUsages(appGuid string) ([]float64, error) {
	promqlResult, err := f.logCacheClient.PromQL(context.Background(), fmt.Sprintf(`absolute_usage{source_id="%s"} / absolute_entitlement{source_id="%s"}`, appGuid, appGuid))
	if err != nil {
		return nil, err
	}

	var instanceUsages []float64
	for _, sample := range promqlResult.GetVector().GetSamples() {
		instanceUsages = append(instanceUsages, sample.GetPoint().GetValue())
	}

	return instanceUsages, nil
}
