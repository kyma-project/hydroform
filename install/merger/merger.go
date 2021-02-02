package merger

const (
	OnConflictLabel   = "on-conflict"
	ReplaceOnConflict = "replace"
)

type Data interface {
	Name() string
	Labels() map[string]string

	LoadOld() (Data, error)
	Update() error
	Merge(old Data) error
}

func Update(secret Data) error {
	if !isReplaceOnConflict(secret.Labels()) {
		err2 := merge(secret)
		if err2 != nil {
			return err2
		}
	}

	err := secret.Update()
	if err != nil {
		return err
	}

	return nil
}

func merge(secret Data) error {
	oldSecret, err := secret.LoadOld()
	if err != nil {
		return err
	}

	return secret.Merge(oldSecret)
}

func isReplaceOnConflict(cm map[string]string) bool {
	val, ok := cm[OnConflictLabel]
	return ok && val == ReplaceOnConflict
}
