package mandrill

import (
	"encoding/base64"
	"fmt"
	"github.com/facebookgo/ensure"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMessageTo(t *testing.T) {
	m := NewMessageTo("testEmail", "testName")
	ensure.DeepEqual(t, m.To[0].Email, "testEmail")
	ensure.DeepEqual(t, m.To[0].Name, "testName")
}

func TestAddRecipient(t *testing.T) {
	m := Message{}
	m.AddRecipient("testEmail", "testName")
	ensure.DeepEqual(t, m.To[0].Email, "testEmail")
	ensure.DeepEqual(t, m.To[0].Name, "testName")
}

func TestAddGlobalMergeVars(t *testing.T) {
	m := Message{}
	gmv := make(map[string]string)
	gmv["testName"] = "testContent"
	m.AddGlobalMergeVars(gmv)
	ensure.DeepEqual(t, "testName", m.GlobalMergeVars[0].Name)
	ensure.DeepEqual(t, "testContent", m.GlobalMergeVars[0].Content)
}

func TestAddTags(t *testing.T) {
	m := Message{}
	m.AddTags("tag1", "tag2")
	ensure.DeepEqual(t, "tag1", m.Tags[0])
	ensure.DeepEqual(t, "tag2", m.Tags[1])
}

func TestAddMetadataField(t *testing.T) {
	m := Message{}
	m.AddMetadataField("testField", "testValue")
	ensure.DeepEqual(t, m.Metadata["testField"], "testValue")
}

func TestAddRecipientMetadata(t *testing.T) {
	m := Message{}
	md := make(map[string]interface{})
	md["testMetaField"] = "testMetaValue"
	m.AddRecipientMetadata("testRecipient1", md)
	ensure.DeepEqual(t, m.RecipientMetadata[0].Recipient, "testRecipient1")
	ensure.DeepEqual(t, m.RecipientMetadata[0].Values["testMetaField"], "testMetaValue")
}

func TestAddSubAccount(t *testing.T) {
	m := Message{}
	m.AddSubAccount("testSubAccount")
	ensure.DeepEqual(t, m.SubAccount, "testSubAccount")
}

func TestAddAttachment(t *testing.T) {
	m := Message{}
	m.AddAttachment([]byte("testData"), "testName", "testMime")
	content, err := base64.StdEncoding.DecodeString(m.Attachments[0].Content)
	ensure.Nil(t, err)
	ensure.DeepEqual(t, []byte("testData"), content)
}

func TestAddHeader(t *testing.T) {
	m := Message{}
	m.AddHeader("testName", "testValue")
	ensure.DeepEqual(t, m.Headers["testName"], "testValue")
}

func TestSend(t *testing.T) {
	srv := httptest.NewServer(&testHandler{
		path:       fmt.Sprintf("test/mandrill/send"),
		respHeader: http.StatusOK,
		respBody:   []byte(`[{"status":"sent","email":"test@test.com","reject_reason": "hard-bounce","_id": "abc123abc123abc123abc123abc123"}]`),
	})
	defer srv.Close()
	m := Message{
		HTML:      "<p> Test HTML </p>",
		Text:      "Test Text",
		Subject:   "Test Subject",
		FromEmail: "test@email.com",
		FromName:  "Test Name",
	}
	m.AddRecipient("userTest@email.com", "test user")
	res, err := m.Send(false, SetMessageUrl(srv.URL))
	ensure.Nil(t, err)
	ensure.DeepEqual(t, res[0].Status, "sent")
	ensure.DeepEqual(t, res[0].Email, "test@test.com")
	ensure.DeepEqual(t, res[0].RejectionReason, "hard-bounce")
	ensure.DeepEqual(t, res[0].Id, "abc123abc123abc123abc123abc123")
}

func TestSendTemplate(t *testing.T) {
	srv := httptest.NewServer(&testHandler{
		path:       fmt.Sprintf("test/mandrill/send"),
		respHeader: http.StatusOK,
		respBody:   []byte(`[{"status":"sent","email":"test@test.com","reject_reason": "hard-bounce","_id": "abc123abc123abc123abc123abc123"}]`),
	})
	defer srv.Close()
	m := Message{
		HTML:      "<p> Test HTML </p>",
		Text:      "Test Text",
		Subject:   "Test Subject",
		FromEmail: "test@email.com",
		FromName:  "Test Name",
	}
	m.AddRecipient("userTest@email.com", "test user")
	content := make(map[string]string)
	content["contentKey"] = "Test Content"
	res, err := m.SendTemplate("test template", content, false, SetMessageUrl(srv.URL))
	ensure.Nil(t, err)
	ensure.DeepEqual(t, res[0].Status, "sent")
	ensure.DeepEqual(t, res[0].Email, "test@test.com")
	ensure.DeepEqual(t, res[0].RejectionReason, "hard-bounce")
	ensure.DeepEqual(t, res[0].Id, "abc123abc123abc123abc123abc123")
}

type testHandler struct {
	path       string
	respHeader int
	respBody   []byte
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(h.respHeader)
	w.Write(h.respBody)
}
