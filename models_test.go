package translatable

import (
	"testing"
)

func TestTranslatableTableName(t *testing.T) {
	var translatable Translatable
	if translatable.TableName() != "translations" {
		t.Errorf("TableName() = %v, want 'translations'", translatable.TableName())
	}
}
