package rabbithole

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Policy unit tests", func() {
	Context("PolicyTarget matching", func() {
		It("MatchesQueues returns true for queue-related targets", func() {
			Ω(PolicyTargetQueues.MatchesQueues()).Should(BeTrue())
			Ω(PolicyTargetClassicQueues.MatchesQueues()).Should(BeTrue())
			Ω(PolicyTargetQuorumQueues.MatchesQueues()).Should(BeTrue())
			Ω(PolicyTargetStreams.MatchesQueues()).Should(BeTrue())
			Ω(PolicyTargetAll.MatchesQueues()).Should(BeTrue())
			Ω(PolicyTargetExchanges.MatchesQueues()).Should(BeFalse())
		})

		It("MatchesExchanges returns true for exchange-related targets", func() {
			Ω(PolicyTargetExchanges.MatchesExchanges()).Should(BeTrue())
			Ω(PolicyTargetAll.MatchesExchanges()).Should(BeTrue())
			Ω(PolicyTargetQueues.MatchesExchanges()).Should(BeFalse())
			Ω(PolicyTargetClassicQueues.MatchesExchanges()).Should(BeFalse())
			Ω(PolicyTargetQuorumQueues.MatchesExchanges()).Should(BeFalse())
			Ω(PolicyTargetStreams.MatchesExchanges()).Should(BeFalse())
		})

		It("MatchesClassicQueues returns true for classic queue targets", func() {
			Ω(PolicyTargetQueues.MatchesClassicQueues()).Should(BeTrue())
			Ω(PolicyTargetClassicQueues.MatchesClassicQueues()).Should(BeTrue())
			Ω(PolicyTargetAll.MatchesClassicQueues()).Should(BeTrue())
			Ω(PolicyTargetQuorumQueues.MatchesClassicQueues()).Should(BeFalse())
			Ω(PolicyTargetStreams.MatchesClassicQueues()).Should(BeFalse())
			Ω(PolicyTargetExchanges.MatchesClassicQueues()).Should(BeFalse())
		})

		It("MatchesQuorumQueues returns true for quorum queue targets", func() {
			Ω(PolicyTargetQueues.MatchesQuorumQueues()).Should(BeTrue())
			Ω(PolicyTargetQuorumQueues.MatchesQuorumQueues()).Should(BeTrue())
			Ω(PolicyTargetAll.MatchesQuorumQueues()).Should(BeTrue())
			Ω(PolicyTargetClassicQueues.MatchesQuorumQueues()).Should(BeFalse())
			Ω(PolicyTargetStreams.MatchesQuorumQueues()).Should(BeFalse())
			Ω(PolicyTargetExchanges.MatchesQuorumQueues()).Should(BeFalse())
		})

		It("MatchesStreams returns true for stream targets", func() {
			Ω(PolicyTargetQueues.MatchesStreams()).Should(BeTrue())
			Ω(PolicyTargetStreams.MatchesStreams()).Should(BeTrue())
			Ω(PolicyTargetAll.MatchesStreams()).Should(BeTrue())
			Ω(PolicyTargetClassicQueues.MatchesStreams()).Should(BeFalse())
			Ω(PolicyTargetQuorumQueues.MatchesStreams()).Should(BeFalse())
			Ω(PolicyTargetExchanges.MatchesStreams()).Should(BeFalse())
		})
	})

	Context("PolicyDefinition CMQ keys detection", func() {
		It("HasCMQKeys returns false for non-CMQ definitions", func() {
			pd := PolicyDefinition{
				"max-length":  1000,
				"message-ttl": 60000,
				"overflow":    "reject-publish",
			}
			Ω(pd.HasCMQKeys()).Should(BeFalse())
		})

		It("HasCMQKeys returns true when ha-mode is present", func() {
			pd := PolicyDefinition{
				"ha-mode":   "exactly",
				"ha-params": 2,
			}
			Ω(pd.HasCMQKeys()).Should(BeTrue())
		})

		It("HasCMQKeys returns true when ha-sync-mode is present", func() {
			pd := PolicyDefinition{
				"ha-mode":      "all",
				"ha-sync-mode": "automatic",
			}
			Ω(pd.HasCMQKeys()).Should(BeTrue())
		})

		It("HasCMQKeys returns true when ha-promote-on-shutdown is present", func() {
			pd := PolicyDefinition{
				"ha-mode":                "all",
				"ha-promote-on-shutdown": "always",
			}
			Ω(pd.HasCMQKeys()).Should(BeTrue())
		})

		It("HasCMQKeys returns false for empty definition", func() {
			pd := PolicyDefinition{}
			Ω(pd.HasCMQKeys()).Should(BeFalse())
		})
	})

	Context("PolicyDefinition WithoutCMQKeys", func() {
		It("removes all CMQ keys", func() {
			pd := PolicyDefinition{
				"ha-mode":                "exactly",
				"ha-params":              2,
				"ha-sync-mode":           "automatic",
				"ha-sync-batch-size":     100,
				"ha-promote-on-shutdown": "always",
				"ha-promote-on-failure":  "when-synced",
				"max-length":             9000,
			}
			result := pd.WithoutCMQKeys()

			Ω(result).Should(HaveLen(1))
			Ω(result).Should(HaveKey("max-length"))
			Ω(result).ShouldNot(HaveKey("ha-mode"))
			Ω(result).ShouldNot(HaveKey("ha-params"))
			Ω(result).ShouldNot(HaveKey("ha-sync-mode"))
			Ω(result).ShouldNot(HaveKey("ha-sync-batch-size"))
			Ω(result).ShouldNot(HaveKey("ha-promote-on-shutdown"))
			Ω(result).ShouldNot(HaveKey("ha-promote-on-failure"))
		})

		It("returns empty definition when only CMQ keys are present", func() {
			pd := PolicyDefinition{
				"ha-mode":   "all",
				"ha-params": 2,
			}
			result := pd.WithoutCMQKeys()
			Ω(result).Should(BeEmpty())
		})

		It("returns same keys when no CMQ keys are present", func() {
			pd := PolicyDefinition{
				"max-length":  1000,
				"message-ttl": 60000,
			}
			result := pd.WithoutCMQKeys()
			Ω(result).Should(HaveLen(2))
			Ω(result).Should(HaveKey("max-length"))
			Ω(result).Should(HaveKey("message-ttl"))
		})

		It("does not modify original definition", func() {
			pd := PolicyDefinition{
				"ha-mode":    "all",
				"max-length": 1000,
			}
			_ = pd.WithoutCMQKeys()
			Ω(pd).Should(HaveLen(2))
			Ω(pd).Should(HaveKey("ha-mode"))
		})
	})

	Context("PolicyDefinition WithoutQuorumQueueIncompatibleKeys", func() {
		It("removes quorum queue incompatible keys", func() {
			pd := PolicyDefinition{
				"ha-mode":              "exactly",
				"max-priority":         10,
				"queue-master-locator": "min-masters",
				"queue-mode":           "lazy",
				"max-length-bytes":     20000000,
			}
			result := pd.WithoutQuorumQueueIncompatibleKeys()

			Ω(result).Should(HaveLen(1))
			Ω(result).Should(HaveKey("max-length-bytes"))
			Ω(result).ShouldNot(HaveKey("ha-mode"))
			Ω(result).ShouldNot(HaveKey("max-priority"))
			Ω(result).ShouldNot(HaveKey("queue-master-locator"))
			Ω(result).ShouldNot(HaveKey("queue-mode"))
		})

		It("keeps quorum-compatible keys", func() {
			pd := PolicyDefinition{
				"max-length":           1000,
				"delivery-limit":       5,
				"dead-letter-exchange": "dlx",
			}
			result := pd.WithoutQuorumQueueIncompatibleKeys()
			Ω(result).Should(HaveLen(3))
		})
	})

	Context("targetsMatch helper function", func() {
		It("PolicyTargetQueues matches Queues and All", func() {
			Ω(targetsMatch(PolicyTargetQueues, PolicyTargetQueues)).Should(BeTrue())
			Ω(targetsMatch(PolicyTargetQueues, PolicyTargetAll)).Should(BeTrue())
			Ω(targetsMatch(PolicyTargetAll, PolicyTargetQueues)).Should(BeTrue())
		})

		It("PolicyTargetExchanges matches Exchanges and All", func() {
			Ω(targetsMatch(PolicyTargetExchanges, PolicyTargetExchanges)).Should(BeTrue())
			Ω(targetsMatch(PolicyTargetExchanges, PolicyTargetAll)).Should(BeTrue())
			Ω(targetsMatch(PolicyTargetAll, PolicyTargetExchanges)).Should(BeTrue())
		})

		It("PolicyTargetQueues does not match Exchanges", func() {
			Ω(targetsMatch(PolicyTargetQueues, PolicyTargetExchanges)).Should(BeFalse())
			Ω(targetsMatch(PolicyTargetExchanges, PolicyTargetQueues)).Should(BeFalse())
		})

		It("Specific queue types match generic Queues target", func() {
			Ω(targetsMatch(PolicyTargetQueues, PolicyTargetClassicQueues)).Should(BeTrue())
			Ω(targetsMatch(PolicyTargetQueues, PolicyTargetQuorumQueues)).Should(BeTrue())
			Ω(targetsMatch(PolicyTargetQueues, PolicyTargetStreams)).Should(BeTrue())
		})
	})

	Context("PolicyDefinition WithoutKeys", func() {
		It("removes specified keys", func() {
			pd := PolicyDefinition{
				"max-age":          "1D",
				"max-length-bytes": 20000000,
				"queue-mode":       "lazy",
			}
			result := pd.WithoutKeys([]string{"max-length-bytes", "queue-mode"})

			Ω(result).Should(HaveLen(1))
			Ω(result).Should(HaveKey("max-age"))
			Ω(result).ShouldNot(HaveKey("max-length-bytes"))
			Ω(result).ShouldNot(HaveKey("queue-mode"))
		})

		It("returns empty when all keys are removed", func() {
			pd := PolicyDefinition{
				"max-age":          "1D",
				"max-length-bytes": 20000000,
			}
			result := pd.WithoutKeys([]string{"max-age", "max-length-bytes", "non-existent"})
			Ω(result).Should(BeEmpty())
		})

		It("handles empty definition", func() {
			pd := PolicyDefinition{}
			result := pd.WithoutKeys([]string{"max-length-bytes"})
			Ω(result).Should(BeEmpty())
		})
	})

	Context("Policy.HasCMQKeys", func() {
		It("returns false for stream policies without CMQ keys", func() {
			p := Policy{
				Name:    "policy.1",
				Vhost:   "/",
				Pattern: "^events",
				ApplyTo: string(PolicyTargetStreams),
				Definition: PolicyDefinition{
					"max-age": "1D",
				},
			}
			Ω(p.HasCMQKeys()).Should(BeFalse())
		})

		It("returns true for policies with CMQ keys", func() {
			p := Policy{
				Name:    "policy.1",
				Vhost:   "/",
				Pattern: "^events",
				ApplyTo: string(PolicyTargetQueues),
				Definition: PolicyDefinition{
					"ha-mode":      "exactly",
					"ha-params":    2,
					"ha-sync-mode": "automatic",
				},
			}
			Ω(p.HasCMQKeys()).Should(BeTrue())
		})

		It("returns false for policies with empty definition", func() {
			p := Policy{
				Name:       "policy.1",
				Vhost:      "/",
				Pattern:    "^events",
				ApplyTo:    string(PolicyTargetQueues),
				Definition: PolicyDefinition{},
			}
			Ω(p.HasCMQKeys()).Should(BeFalse())
		})
	})

	Context("Policy.DoesMatchName", func() {
		It("matches queue names with simple pattern", func() {
			p := Policy{
				Name:    "policy.1",
				Vhost:   "/",
				Pattern: "^events",
				ApplyTo: string(PolicyTargetQueues),
				Definition: PolicyDefinition{
					"max-length": 100000,
				},
			}
			Ω(p.DoesMatchName("/", "events.1", PolicyTargetQueues)).Should(BeTrue())
		})

		It("matches with complex regex patterns", func() {
			p := Policy{
				Name:    "policy.2",
				Vhost:   "/",
				Pattern: `^ca\.`,
				ApplyTo: string(PolicyTargetQueues),
				Definition: PolicyDefinition{
					"max-length": 100000,
				},
			}
			Ω(p.DoesMatchName("/", "ca.on.to.1", PolicyTargetQueues)).Should(BeTrue())
			Ω(p.DoesMatchName("/", "cdi.r.1", PolicyTargetQueues)).Should(BeFalse())
			Ω(p.DoesMatchName("/", "ca", PolicyTargetQueues)).Should(BeFalse())
			Ω(p.DoesMatchName("/", "abc.r.1", PolicyTargetQueues)).Should(BeFalse())
		})

		It("does not match different vhosts", func() {
			p := Policy{
				Name:    "policy.2",
				Vhost:   "/",
				Pattern: `^ca\.`,
				ApplyTo: string(PolicyTargetQueues),
				Definition: PolicyDefinition{
					"max-length": 100000,
				},
			}
			Ω(p.DoesMatchName("other-vhost", "ca.on.to.2", PolicyTargetQueues)).Should(BeFalse())
		})

		It("respects target types", func() {
			p := Policy{
				Name:    "policy.3",
				Vhost:   "/",
				Pattern: `^events\.`,
				ApplyTo: string(PolicyTargetExchanges),
				Definition: PolicyDefinition{
					"alternate-exchange": "amq.fanout",
				},
			}
			Ω(p.DoesMatchName("/", "events.regional.na", PolicyTargetExchanges)).Should(BeTrue())
			Ω(p.DoesMatchName("/", "events.regional.na", PolicyTargetQueues)).Should(BeFalse())
			Ω(p.DoesMatchName("/", "events.regional.na", PolicyTargetClassicQueues)).Should(BeFalse())
			Ω(p.DoesMatchName("/", "events.regional.na", PolicyTargetStreams)).Should(BeFalse())
		})
	})

	Context("OperatorPolicy.HasCMQKeys", func() {
		It("returns true for policies with CMQ keys", func() {
			p := OperatorPolicy{
				Name:    "op-policy.1",
				Vhost:   "/",
				Pattern: "^queues",
				ApplyTo: string(PolicyTargetQueues),
				Definition: PolicyDefinition{
					"ha-mode": "all",
				},
			}
			Ω(p.HasCMQKeys()).Should(BeTrue())
		})

		It("returns false for policies without CMQ keys", func() {
			p := OperatorPolicy{
				Name:    "op-policy.1",
				Vhost:   "/",
				Pattern: "^queues",
				ApplyTo: string(PolicyTargetQueues),
				Definition: PolicyDefinition{
					"max-length": 1000,
				},
			}
			Ω(p.HasCMQKeys()).Should(BeFalse())
		})
	})

	Context("OperatorPolicy.DoesMatchName", func() {
		It("matches queue names", func() {
			p := OperatorPolicy{
				Name:    "op-policy.1",
				Vhost:   "/",
				Pattern: "^orders",
				ApplyTo: string(PolicyTargetQueues),
				Definition: PolicyDefinition{
					"max-length": 5000,
				},
			}
			Ω(p.DoesMatchName("/", "orders.pending", PolicyTargetQueues)).Should(BeTrue())
			Ω(p.DoesMatchName("/", "invoices.pending", PolicyTargetQueues)).Should(BeFalse())
		})
	})
})
