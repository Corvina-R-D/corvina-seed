package utils_test

import (
	"corvina/corvina-seed/src/utils"
	"testing"
)

func TestIcanGenerateRandomName(t *testing.T) {
	name := utils.RandomName()

	if name == "" {
		t.Error("Name is empty")
	}
}
