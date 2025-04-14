package story

type Story struct {
	storage stg
}

func New(storage stg) *Story {
	return &Story{
		storage: storage,
	}
}
