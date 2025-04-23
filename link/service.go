package link

type Repository interface {
	Get(id string) (link string, err error)
	Add(link string) (id string, err error)
}

type Service struct {
	Repository
}

func NewService(linkRepo Repository) *Service {
	return &Service{linkRepo}
}
