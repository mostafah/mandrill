// Package mandrill gives a simple interface for sending email through
// Mandrill's API, documented at https://mandrillapp.com/api/docs/.
//
// This is not a full implementation of the API and only provides few essential
// sending out emails.
//
// To use this package first set your API key:
//
//     mandrill.Key = "xxxx"
//     // you can test your API key with Ping
//     err := mandrill.Ping()
//     // everything is OK if err is nil
//
//
// It's easy to send a message:
//
//     msg := mandrill.NewMessageTo("recipient@domain.com", "recipient's name")
//     msg.HTML = "HTML content"
//     msg.Text = "plain text content" // optional
//     msg.Subject = "subject"
//     msg.FromEmail = "email@domain.com"
//     msg.FromName = "your name"
//     res, err := msg.Send(fase)
//
// It's even easier to send a message using a template:
//
//     res, err := mandrill.NewMessageTo(email, name).SendTemplate(tmplName, data, false)
package mandrill

import (
	"encoding/base64"
	"fmt"
	"github.com/jmcvetta/napping"
)

// API key for Mandrill user. You should set this to your API key before calling
// any of the functions. You can get a API key for your account in your
// Mandrill account settings.
var Key string

// type Error holds error return messages from API calls.
type Error struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// type Attachment holds necessary information for an attachment
// Mime - the MIME type of the attachment
// Name - the name of the attachment
// Content - Base64 encoded string of file content
type Attachment struct {
	Mime    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// newError returns a new Error instance.
func newError() *Error {
	return &Error{}
}

// Error porduces error message for err.
func (err *Error) Error() string {
	return fmt.Sprintf("mandrill: %s: %s", err.Name, err.Message)
}

// do is an easy function for performing requests against Mandrill's API.
func do(url string, data interface{}, result interface{}) error {
	err := newError()

	rr := &napping.Request{
		Url:    "https://mandrillapp.com/api/1.0" + url,
		Method: "POST",
		Payload:   data,
		Result: result,
		Error:  err}

	status, errs := napping.Send(rr)

	if errs == nil {
		return nil
	}

	fmt.Println(status, rr.Error)

	return errs
}

// Ping validates your API key. Call this to make sure your API key is correct.
// It should return nil as error if everything is OK.
func Ping() error {
	var data struct {
		Key string `json:"key"`
	}
	data.Key = Key
	return do("/users/ping", &data, nil)
}

// Type SendResult holds information returned by send requests.
type SendResult struct {
	// email address of the recipient
	Email string `json:"email"`
	// the sending status
	// either "sent", "queued", "rejected", or "invalid"
	Status string `json:"status"`
	// the reason for rejection if status is "rejected"
	RejectionReason string `json:"reject_reason"`
	// the message's unique id
	Id string `json:"_id"`
}

// Type To holds information about a recipient for a message.
type To struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type RecipientMetadata struct {
	Recipient string                 `json:"rcpt"`
	Values    map[string]interface{} `json:"values"`
}

// Type Message represents an email message for Mandrill.
type Message struct {
	// full HTML content to be sent
	HTML string `json:"html,omitempty"`
	// full plain text content to be sent
	Text string `json:"text,omitempty"`
	// the message subject
	Subject string `json:"subject,omitempty"`
	// the sender email address
	FromEmail string `json:"from_email,omitempty"`
	// name of the sender
	FromName string `json:"from_name,omitempty"`
	// recipient(s) information
	To []*To `json:"to"`
	// global merge variables to use for all recipients
	GlobalMergeVars []*variable `json:"global_merge_vars,omitempty"`
	// Mandrill tags
	Tags []string `json:"tags,omitempty"`
	// message message metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// per recipient metadata
	RecipientMetadata []*RecipientMetadata `json:"recipient_metadata,omitempty"`
	// the subaccount name to use
	SubAccount string `json:"subaccount,omitempty"`
	// attachments
	Attachments []*Attachment `json:"attachments,omitempty"`
	// TODO implement other fields
}

// NewMessage returns a new instance of Message.
func NewMessage() *Message {
	return &Message{}
}

// NewMessageTo makes a new message with specified recipient.
func NewMessageTo(email, name string) *Message {
	return NewMessage().AddRecipient(email, name)
}

// AddRecipient adds a new recpipeint for msg.
func (msg *Message) AddRecipient(email, name string) *Message {
	to := &To{email, name}
	msg.To = append(msg.To, to)
	return msg
}

// AddGlobalMergeVars provides given data as merge vars with message.
func (msg *Message) AddGlobalMergeVars(data map[string]string) *Message {
	msg.GlobalMergeVars = append(msg.GlobalMergeVars, mapToVars(data)...)
	return msg
}

// AddTags does what it's name says.
func (msg *Message) AddTags(tags ...string) *Message {
	msg.Tags = append(msg.Tags, tags...)
	return msg
}

// AddMetadataField adds a new field/value to the metadata.
func (msg *Message) AddMetadataField(field string, value interface{}) *Message {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]interface{})
	}
	msg.Metadata[field] = value
	return msg
}

// AddRecipientMetadata adds a new recipient to the recipient metadata.
func (msg *Message) AddRecipientMetadata(recipient string, metadata map[string]interface{}) *Message {
	rcptMetadata := &RecipientMetadata{recipient, metadata}
	msg.RecipientMetadata = append(msg.RecipientMetadata, rcptMetadata)
	return msg
}

// AddSubAccount will set the subaccount for the message to be delivered by.
func (msg *Message) AddSubAccount(subaccount string) *Message {
	msg.SubAccount = subaccount
	return msg
}

// AddAttachment will add an attachment to be sent via Mandrill.
// Function accepts a single attachment that contains
// a file name a data byte slice and a mime type, from that content is
// set using encoding/base64.
// This function may be called multiple times to add additional attachments.
func (msg *Message) AddAttachment(data []byte, name, mime string) *Message {
	attachment := &Attachment{
		Mime:    mime,
		Name:    name,
		Content: base64.StdEncoding.EncodeToString(data),
	}

	msg.Attachments = append(msg.Attachments, attachment)
	return msg
}

// Send performs a send request for msg.
func (msg *Message) Send(async bool) ([]*SendResult, error) {
	// prepare request data
	var data struct {
		Key     string   `json:"key"`
		Message *Message `json:"message,omitempty"`
		Async   bool     `json:"async"`
	}
	data.Key = Key
	data.Message = msg
	data.Async = async

	// perform the request
	res := make([]*SendResult, 0)
	err := do("/messages/send", &data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// SendTemplate performs a template-based send request for msg.
func (msg *Message) SendTemplate(tmpl string, content map[string]string, async bool) ([]*SendResult, error) {
	// prepare request data
	var data struct {
		Key            string      `json:"key"`
		TemplateName   string      `json:"template_name"`
		TemplateConent []*variable `json:"template_content"`
		Message        *Message    `json:"message,omitempty"`
		Async          bool        `json:"async"`
	}
	data.Key = Key
	data.TemplateName = tmpl
	data.TemplateConent = mapToVars(content)
	data.Message = msg
	data.Async = async

	// perform the request
	res := make([]*SendResult, 0)
	err := do("/messages/send-template", &data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Type variable holds one piece of data for dynamic content of messages.
type variable struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// mapToVars converts a map to a list variable.
func mapToVars(m map[string]string) []*variable {
	vars := make([]*variable, 0, len(m))
	for k, v := range m {
		vars = append(vars, &variable{k, v})
	}
	return vars
}
