package db

import (
	"reflect"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestGetCoalesceAssignmentByDialect(t *testing.T) {
	tests := []struct {
		name      string
		dialect   string
		tableName string
		columns   []string
		want      clause.Set
	}{
		{
			name:      "mysql uses VALUES()",
			dialect:   "mysql",
			tableName: "t_user",
			columns:   []string{"address"},
			want: clause.Assignments(map[string]interface{}{
				"address": gorm.Expr("COALESCE(VALUES(address), address)"),
			}),
		},
		{
			name:      "dameng uses excluded",
			dialect:   "dameng",
			tableName: "t_user",
			columns:   []string{"address"},
			want: clause.Assignments(map[string]interface{}{
				"address": gorm.Expr("COALESCE(excluded.address, t_user.address)"),
			}),
		},
		{
			name:      "unsupported dialect returns empty set",
			dialect:   "postgres",
			tableName: "t_user",
			columns:   []string{"address"},
			want:      clause.Set{},
		},
		{
			name:      "no columns returns empty set",
			dialect:   "mysql",
			tableName: "t_user",
			columns:   nil,
			want:      clause.Set{},
		},
		{
			name:      "no table name returns empty set",
			dialect:   "mysql",
			tableName: "",
			columns:   []string{"address"},
			want:      clause.Set{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCoalesceAssignmentByDialect(tt.dialect, tt.tableName, tt.columns...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCoalesceAssignmentByDialect(%q, %q) = %v, want %v", tt.dialect, tt.tableName, got, tt.want)
			}
		})
	}
}
