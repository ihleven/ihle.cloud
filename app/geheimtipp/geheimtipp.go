package geheimitpp

import "github.com/interhome-group/cms/content"

func NewService(repo dbrepo) (*Service, error) {

	return &Service{Repo: repo}, nil
}

type Service struct {
	Repo dbrepo
}

type Something struct{}

// implementiert in pkg db
type dbrepo interface {
	LoadSomething(someparam int) (*Something, error)
	StoreSomething(s *Something) error
}

type Edition struct {
	content.EntryContent `type:"Edition" json:"-" yaml:"-" mimetype:"text/json"`

	Key        string `json:"key"`
	Geburtstag string `json:"geburtstag"`
	Todestag   string `json:"todestag"`
	Vater      string `json:"vater"`
	Mutter     string `json:"mutter"`

	Content string `json:"markdown" yaml:"-"`
}
