package store

// Repository abstracts persisted state storage so Store logic does not depend on a specific backend format.
type Repository interface {
	Load() (PersistedState, bool, error)
	Save(state PersistedState) error
	Path() string
}
