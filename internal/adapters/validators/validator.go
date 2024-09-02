package validators

import (
	"errors"
	"strings"
)

type Repo string

func (r *Repo) Validate() error {
	repoSlice := strings.Split(string(*r), "/")
	if len(repoSlice) < 2 || len(repoSlice) >= 3 {
		return errors.New("invalid repo")
	}

	return nil
}
