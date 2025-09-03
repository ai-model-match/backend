package rsEngine

type rsEngineRepositoryInterface interface {
}

type rsEngineRepository struct {
}

func newRsEngineRepository() rsEngineRepository {
	return rsEngineRepository{}
}
