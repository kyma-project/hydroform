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
	if isMerge(data) {
		err := data.Merge()
		if err != nil {
			return err
		}
	}

	return data.Update()
}

func isMerge(data Data) bool {
	labels := data.Labels()
	if labels == nil {
		return true
	}

	val, ok := (*labels)[OnConflictLabel]
	return !ok || val != ReplaceOnConflict
}
