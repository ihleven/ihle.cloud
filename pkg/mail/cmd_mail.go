package mail

import (
	"fmt"
	"net/smtp"
	"net/url"

	"bitbucket.org/hotelplan/webcc-content/cms/pkg/errors"
)

type MailCmd struct {
	SMTPAddr string `arg:"--smtp-addr,env:SMTP_ADDR" default:"smtp.de.opalstack.com:587"`
	SMTPAuth string `arg:"--smtp-auth,env:SMTP_AUTH" default:"plain://username:password@host"`
	SMTPFrom string `arg:"--smtp-from,env:SMTP_FROM" default:"email@address"`

	Receiver string `arg:"--receiver"                default:"matthias@ihle.cloud"`
}

func (cmd *MailCmd) Run() error {
	fmt.Println("mail:", cmd)
	to := []string{cmd.Receiver}
	msg := []byte("To:MAtthias Ihle <cloud@ihleven.de>\nFrom:bar@ihleven.de\nSubject: Hello World\n\nDas ist der Body")
	auth, err := smtpauth(cmd.SMTPAuth)
	if err != nil {
		return err
	}
	err = smtp.SendMail(cmd.SMTPAddr, auth, cmd.SMTPFrom, to, msg)
	if err != nil {
		return err
	}
	return nil
}

func smtpauth(urlconf string) (smtp.Auth, error) {
	url, err := url.Parse(urlconf)
	if err != nil {
		return nil, errors.New("invalid mail auth config: %s", urlconf)
	}
	switch url.Scheme {
	case "plain":
		pwd, _ := url.User.Password()
		auth := smtp.PlainAuth("", url.User.Username(), pwd, url.Host)
		return auth, nil
	}
	return nil, errors.New("unsupported mail auth scheme: %s", url.Scheme)
}
