package test

import (
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

func TestListSessions(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestListSessionsWithPersonaFilter(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions?personaId=1").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetSession(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/ddd123").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetSessionMessages(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/784bf78750994cddba767f950ba0596a/messages").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestDeleteSession(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/ddd123").
		Delete().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}

func TestGetSessionNotFound(t *testing.T) {
	str, httpResp := requests.NewHTTPRequest(domain+"/api/sessions/nonexistent-session-id").
		Get().
		SetHeaderValue("Authorization", "Bearer "+token).
		ToString()
	t.Log(str, httpResp)
}
