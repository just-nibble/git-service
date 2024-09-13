package validator

import "strings"

func IsRepository(repoName string) bool {
	return strings.Contains(repoName, "/")
}
