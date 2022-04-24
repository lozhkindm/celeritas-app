package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"time"

	apimail "github.com/ainsleyclark/go-mail"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain       string
	TemplatesDir string
	Host         string
	Port         int
	Username     string
	Password     string
	Encryption   string
	FromAddress  string
	FromName     string
	Jobs         chan Message
	Results      chan Result
	API          string
	APIKey       string
	APIUrl       string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

type Result struct {
	Success bool
	Error   error
}

func (m *Mail) ListenForMail() {
	for {
		msg := <-m.Jobs
		err := m.Send(msg)
		m.Results <- Result{Success: err == nil, Error: err}
	}
}

func (m *Mail) Send(msg Message) error {
	if len(m.API) > 0 && len(m.APIKey) > 0 && len(m.APIUrl) > 0 && m.API != "smtp" {
		return m.selectAPI(msg)
	}
	return m.SendSMTPMessage(msg)
}

func (m *Mail) SendViaAPI(msg Message, transport string) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	cfg := apimail.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	driver, err := apimail.NewClient(transport, cfg)
	if err != nil {
		return err
	}

	html, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}
	plain, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	tx := &apimail.Transmission{
		Recipients: []string{msg.To},
		Subject:    msg.Subject,
		HTML:       html,
		PlainText:  plain,
	}

	if err := m.addAPIAttachments(msg, tx); err != nil {
		return err
	}
	if _, err := driver.Send(tx); err != nil {
		return err
	}

	return nil
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	htmlMsg, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMsg, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.
		SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject).
		SetBody(mail.TextHTML, htmlMsg).
		AddAlternative(mail.TextPlain, plainMsg)

	if len(msg.Attachments) > 0 {
		for _, a := range msg.Attachments {
			email.AddAttachment(a)
		}
	}

	return email.Send(client)
}

func (m *Mail) selectAPI(msg Message) error {
	switch m.API {
	case "mailgun", "sparkpost", "sendgrid":
		return m.SendViaAPI(msg, m.API)
	default:
		return fmt.Errorf("unknown api %q", m.API)
	}
}

func (m *Mail) addAPIAttachments(msg Message, tx *apimail.Transmission) error {
	if len(msg.Attachments) > 0 {
		attachments := make([]apimail.Attachment, len(msg.Attachments))
		for _, a := range msg.Attachments {
			bts, err := ioutil.ReadFile(a)
			if err != nil {
				return err
			}

			att := apimail.Attachment{
				Filename: filepath.Base(a),
				Bytes:    bts,
			}
			attachments = append(attachments, att)
		}
		tx.Attachments = attachments
	}
	return nil
}

func (m *Mail) getEncryption(enc string) mail.Encryption {
	switch enc {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	tmpl := fmt.Sprintf("%s/%s.html.tmpl", m.TemplatesDir, msg.Template)
	t, err := template.New("email-html").ParseFiles(tmpl)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	return m.inlineCSS(tpl.String())
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	tmpl := fmt.Sprintf("%s/%s.plain.tmpl", m.TemplatesDir, msg.Template)
	t, err := template.New("email-plain").ParseFiles(tmpl)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func (m *Mail) inlineCSS(doc string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	p, err := premailer.NewPremailerFromString(doc, &options)
	if err != nil {
		return "", err
	}

	return p.Transform()
}
