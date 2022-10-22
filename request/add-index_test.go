package request

import "testing"

func TestMergeSql(t *testing.T) {
	type Case struct {
		tableName string
		colName   []string
		expected  string
	}
	cases := []Case{
		Case{
			"test1",
			[]string{"a", "b"},
			"alter table test1 add index AdvisorIndex1(a,b);",
		},
		Case{
			"test2",
			[]string{"b", "c"},
			"alter table test2 add index AdvisorIndex2(b,c);",
		},
	}
	for _, c := range cases {
		actual := MergeSql(c.tableName, c.colName)
		if actual != c.expected {
			t.Errorf("expected %v, but got %v", c.expected, actual)
		}
	}
}
