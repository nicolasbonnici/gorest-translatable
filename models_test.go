package translatable

import (
	"testing"
)

func TestTranslatableTableName(t *testing.T) {
	var translatable Translatable
	if translatable.TableName() != "translatable" {
		t.Errorf("TableName() = %v, want 'translatable'", translatable.TableName())
	}
}
