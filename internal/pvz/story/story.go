package story

type Story struct {
	storage   stg
	txManager manager
}

func New(storage stg, txManager manager) *Story {
	return &Story{
		storage:   storage,
		txManager: txManager,
	}
}
