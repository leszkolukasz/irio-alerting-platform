package email

type MockMailer struct {
	SendCalled     bool
	LastTo         string
	LastIncidentID string
	LastServiceID  uint64
	Err            error
}

func (m *MockMailer) SendNotification(toEmail string, incidentID string, serviceID uint64) error {
	m.SendCalled = true
	m.LastTo = toEmail
	m.LastIncidentID = incidentID
	m.LastServiceID = serviceID
	return m.Err
}
