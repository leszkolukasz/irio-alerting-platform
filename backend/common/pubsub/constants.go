package pubsub

const (
	ServiceUpTopic                  = "service-up"
	ServiceDownTopic                = "service-down"
	ServiceCreatedTopic             = "service-created"
	ServiceModifiedTopic            = "service-modified"
	ServiceRemovedTopic             = "service-removed"
	IncidentStartTopic              = "incident-start"
	IncidentResolvedTopic           = "incident-resolved"
	IncidentAcknowledgeTimeoutTopic = "incident-acknowledge-timeout"
	IncidentUnresolvedTopic         = "incident-unresolved"
	NotifyOncallerTopic             = "notify-oncaller"
	OncallerAcknowledgedTopic       = "oncaller-acknowledged"
	ExecuteHealthCheckTopic         = "execute-health-check"
)
