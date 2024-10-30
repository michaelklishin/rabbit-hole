package rabbithole

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deprecated feature flags", func() {
	testFunc := func(b []byte, expectedPhase DeprecationPhase, wantErr bool) {
		var d DeprecationPhase
		if wantErr {
			Ω(d.UnmarshalJSON(b)).To(MatchError(ContainSubstring("unknown deprecation phase:")))
			return
		}
		Ω(d.UnmarshalJSON(b)).To(Succeed())
		Ω(d).To(Equal(expectedPhase))
	}
	DescribeTable("deprecation phase unmarshal and mapping", testFunc,
		Entry("Parsing permitted by default", []byte(`"permitted_by_default"`), DeprecationPermittedByDefault, false),
		Entry("Parsing denied by default", []byte(`"denied_by_default"`), DeprecationDeniedByDefault, false),
		Entry("Parsing disconnect", []byte(`"disconnect"`), DeprecationDisconnected, false),
		Entry("Parsing removed", []byte(`"removed"`), DeprecationRemoved, false),
		Entry("Parsing an invalid state", []byte(`"none-of-the-above"`), DeprecationPhase(-1), true),
	)
})
