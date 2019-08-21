package reporter // import "code.cloudfoundry.org/cpu-entitlement-admin-plugin/reporter"

import (
	"code.cloudfoundry.org/cli/plugin"
	plugin_models "code.cloudfoundry.org/cli/plugin/models"
)

type Report struct {
	SpaceReports []SpaceReport
}

type SpaceReport struct {
	SpaceName string
	Apps      []string
}

//go:generate counterfeiter . MetricsFetcher

type MetricsFetcher interface {
	FetchInstanceEntitlementUsages(appGuid string) ([]float64, error)
}

// type CloudFoundryClient interface {
// 	GetSpaces() []Space
// }

// type Space struct {
// 	Name         string
// 	Applications []Application
// }

// type Application struct {
// 	Name string
// 	Guid string
// }

type Reporter struct {
	cli            plugin.CliConnection
	metricsFetcher MetricsFetcher
}

func New(cli plugin.CliConnection, metricsFetcher MetricsFetcher) Reporter {
	return Reporter{
		cli:            cli,
		metricsFetcher: metricsFetcher,
	}
}

func (r Reporter) OverEntitlementInstances() (Report, error) {
	spaceReports := []SpaceReport{}

	spaces, _ := r.cli.GetSpaces()
	for _, space := range spaces {
		spaceModel, err := r.cli.GetSpace(space.Name)
		if err != nil {
			return Report{}, err
		}

		apps, err := r.filterApps(spaceModel.Applications)
		if err != nil {
			return Report{}, err
		}

		if len(apps) == 0 {
			continue
		}

		spaceReports = append(spaceReports, SpaceReport{SpaceName: space.Name, Apps: apps})
	}

	return Report{SpaceReports: spaceReports}, nil
}

func (r Reporter) filterApps(spaceApps []plugin_models.GetSpace_Apps) ([]string, error) {
	apps := []string{}
	for _, app := range spaceApps {
		isOverEntitlement, err := r.isOverEntitlement(app.Guid)
		if err != nil {
			return nil, err
		}
		if isOverEntitlement {
			apps = append(apps, app.Name)
		}
	}
	return apps, nil
}

func (r Reporter) isOverEntitlement(appGuid string) (bool, error) {
	appInstancesUsages, err := r.metricsFetcher.FetchInstanceEntitlementUsages(appGuid)
	if err != nil {
		return false, err
	}

	isOverEntitlement := false
	for _, usage := range appInstancesUsages {
		if usage > 1 {
			isOverEntitlement = true
		}
	}

	return isOverEntitlement, nil
}
