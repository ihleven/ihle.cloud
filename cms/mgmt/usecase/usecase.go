package usecase

import (
	"sync"

	"bitbucket.org/hotelplan/webcc-content/cms/mgmt/gitrepo"
)

type Service struct {
	//  //
	// //
	////
	///
	//
	sync.RWMutex
	repo *gitrepo.Sitory
	// CMS  *mgmt.CMS
	// Next bool
}

func New(repo *gitrepo.Sitory, next bool) (*Service, error) {

	return &Service{repo: repo}, nil
	// CMS: cms, Next: next
}

func (s *Service) Repo() *gitrepo.Sitory {
	return s.repo
}
