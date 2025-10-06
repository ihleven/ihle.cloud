package geheimitpp

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
