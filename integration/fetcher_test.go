package integration_test

import (
	"strings"

	"code.cloudfoundry.org/cpu-entitlement-admin-plugin/metrics"
	. "code.cloudfoundry.org/cpu-entitlement-admin-plugin/test_utils"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Fetcher", func() {
	var (
		org string
		uid string

		logCacheFetcher metrics.LogCacheFetcher
	)

	BeforeEach(func() {
		uid = uuid.New().String()
		org = "org-" + uid
		space := "space-" + uid

		Expect(Cmd("cf", "create-org", org).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "target", "-o", org).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "create-space", space).Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "target", "-o", org, "-s", space).Run()).To(gexec.Exit(0))

		Expect(Cmd("cf", "push", "spinner-1-"+uid, "-i", "3").WithDir("../../spinner").WithTimeout("2m").Run()).To(gexec.Exit(0))
		Expect(Cmd("cf", "push", "spinner-2-"+uid).WithDir("../../spinner").WithTimeout("2m").Run()).To(gexec.Exit(0))

		logCacheURL := strings.Replace(cfApi, "https://api.", "http://log-cache.", 1)
		getToken := func() (string, error) {
			return getCmdOutput("cf", "oauth-token"), nil
		}

		logCacheFetcher = metrics.NewLogCacheFetcher(logCacheURL, getToken)
	})

	AfterEach(func() {
		Expect(Cmd("cf", "delete-org", "-f", org).WithTimeout("1m").Run()).To(gexec.Exit(0))
	})

	It("prints the list of apps that are over entitlement", func() {
		getUsages := func(appName string) []float64 {
			appGuid := getCmdOutput("cf", "app", appName, "--guid")
			usages, err := logCacheFetcher.FetchInstanceEntitlementUsages(appGuid)
			Expect(err).NotTo(HaveOccurred())
			return usages
		}

		Eventually(getUsages("spinner-1-"+uid), "20s", "1s").Should(HaveLen(3))
		Eventually(getUsages("spinner-2-"+uid), "20s", "1s").Should(HaveLen(1))
	})

	// invalid log-cache url
	// non-existant app
	// app with no instances
})

func getCmdOutput(cmd string, args ...string) string {
	return strings.TrimSpace(string(Cmd(cmd, args...).Run().Out.Contents()))
}
