package reporter_test

import (
	"errors"
	"fmt"

	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	reporterpkg "code.cloudfoundry.org/cpu-entitlement-admin-plugin/reporter"
	"code.cloudfoundry.org/cpu-entitlement-admin-plugin/reporter/reporterfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Reporter", func() {
	var (
		reporter           reporterpkg.Reporter
		fakeCliConnection  *pluginfakes.FakeCliConnection
		fakeMetricsFetcher *reporterfakes.FakeMetricsFetcher
	)

	BeforeEach(func() {
		fakeCliConnection = new(pluginfakes.FakeCliConnection)
		fakeMetricsFetcher = new(reporterfakes.FakeMetricsFetcher)

		fakeCliConnection.GetSpacesReturns([]plugin_models.GetSpaces_Model{
			{Guid: "space1-guid", Name: "space1"},
			{Guid: "space2-guid", Name: "space2"},
		}, nil)

		fakeCliConnection.GetSpaceStub = func(spaceName string) (plugin_models.GetSpace_Model, error) {
			switch spaceName {
			case "space1":
				return plugin_models.GetSpace_Model{
					Applications: []plugin_models.GetSpace_Apps{
						{Name: "app1", Guid: "space1-app1-guid"},
						{Name: "app2", Guid: "space1-app2-guid"},
					},
				}, nil
			case "space2":
				return plugin_models.GetSpace_Model{
					Applications: []plugin_models.GetSpace_Apps{
						{Name: "app1", Guid: "space2-app1-guid"},
					},
				}, nil
			}

			return plugin_models.GetSpace_Model{}, fmt.Errorf("Space '%s' not found", spaceName)
		}

		fakeMetricsFetcher.FetchInstanceEntitlementUsagesStub = func(appGuid string) ([]float64, error) {
			switch appGuid {
			case "space1-app1-guid":
				return []float64{1.5, 0.5}, nil
			case "space1-app2-guid":
				return []float64{0.3}, nil
			case "space2-app1-guid":
				return []float64{0.2}, nil
			}

			return nil, nil
		}

		reporter = reporterpkg.New(fakeCliConnection, fakeMetricsFetcher)
	})

	Describe("OverEntitlementInstances", func() {
		var (
			report reporterpkg.Report
			err    error
		)

		JustBeforeEach(func() {
			report, err = reporter.OverEntitlementInstances()
		})

		It("succeeds", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns all instances that are over entitlement", func() {
			Expect(report).To(Equal(reporterpkg.Report{
				SpaceReports: []reporterpkg.SpaceReport{
					reporterpkg.SpaceReport{
						SpaceName: "space1",
						Apps: []string{
							"app1",
						},
					},
				},
			}))
		})

		When("fetching the list of apps fails", func() {
			BeforeEach(func() {
				fakeCliConnection.GetSpaceReturns(plugin_models.GetSpace_Model{}, errors.New("get-space-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("get-space-error"))
			})
		})

		When("getting the entitlement usage for an app fails", func() {
			BeforeEach(func() {
				fakeMetricsFetcher.FetchInstanceEntitlementUsagesReturns(nil, errors.New("fetch-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("fetch-error"))
			})
		})
	})
})
