package merger

const (
	OnConflictLabel   string = "on-conflict"
	ReplaceOnConflict string = "replace"
)

type Data interface {
	Labels() *map[string]string
	Update() error
	Merge() error
}

func Update(data Data) error {
	if !isReplaceOnConflict(*data.Labels()) {
		err := data.Merge()
		if err != nil {
			return err
		}
	}

	err := data.Update()
	if err != nil {
		return err
	}

	return nil
}

func isReplaceOnConflict(labels map[string]string) bool {
	val, ok := labels[OnConflictLabel]
	return ok && val == ReplaceOnConflict
}
