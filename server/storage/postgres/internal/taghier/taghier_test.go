// +build all_tests unit_tests

package taghier

import (
	"reflect"
	"sort"
	"testing"

	"github.com/juju/errors"
)

// Hierarchy is as follows:
//
// ├── 1
// │   ├── 4
// │   │   └── 7
// │   │       └── 8
// │   ├── 5
// │   │   └── 9
// │   └── 6
// │       └── 10
// │           ├── 11
// │           └── 12
// ├── 2
// │   └── 13
// │       ├── 14
// │       └── 15
// └── 3
//     └── 16
type tmpRegistry struct{}

func (tr *tmpRegistry) GetParent(id int) (int, error) {
	switch id {
	case 0:
		panic("zero id is illegal")
	case 1:
		return 0, nil
	case 2:
		return 0, nil
	case 3:
		return 0, nil

	case 4:
		return 1, nil
	case 5:
		return 1, nil
	case 6:
		return 1, nil

	case 7:
		return 4, nil

	case 8:
		return 7, nil

	case 9:
		return 5, nil

	case 10:
		return 6, nil

	case 11:
		return 10, nil
	case 12:
		return 10, nil

	case 13:
		return 2, nil

	case 14:
		return 13, nil
	case 15:
		return 13, nil

	case 16:
		return 3, nil
	}

	panic("no tag")
}

func TestHier(t *testing.T) {
	reg := tmpRegistry{}
	hier := New(&reg)
	hier.Add(4)
	if err := check(hier, []int{4}, []int{1, 4}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	hier.Add(7)
	if err := check(hier, []int{7}, []int{1, 4, 7}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	hier.Add(1)
	if err := check(hier, []int{7}, []int{1, 4, 7}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	hier.Add(7)
	if err := check(hier, []int{7}, []int{1, 4, 7}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	hier.Add(5)
	if err := check(hier, []int{5, 7}, []int{1, 4, 5, 7}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	hier.Add(12)
	if err := check(hier, []int{5, 7, 12}, []int{1, 4, 5, 6, 7, 10, 12}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	hier.Add(9)
	if err := check(hier, []int{7, 9, 12}, []int{1, 4, 5, 6, 7, 9, 10, 12}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	hier.Add(3)
	if err := check(hier, []int{3, 7, 9, 12}, []int{1, 3, 4, 5, 6, 7, 9, 10, 12}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}
}

func check(hier *TagHier, leafs, all []int) error {
	if !reflect.DeepEqual(hier.GetLeafs(), leafs) {
		return errors.Errorf("leafs are wrong: expected %v, got %v", leafs, hier.GetLeafs())
	}

	if !reflect.DeepEqual(hier.GetAll(), all) {
		return errors.Errorf("all tags are wrong: expected %v, got %v", all, hier.GetAll())
	}

	return nil
}

func TestDiff(t *testing.T) {
	if err := checkDiff([]int{}, []int{1, 3, 4}, []int{1, 3, 4}, []int{}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	if err := checkDiff([]int{1, 4}, []int{1, 3, 4}, []int{3}, []int{}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	if err := checkDiff([]int{1, 4, 7, 9}, []int{1, 3, 4}, []int{3}, []int{7, 9}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}

	if err := checkDiff([]int{1, 4, 7, 9}, []int{}, []int{}, []int{1, 4, 7, 9}); err != nil {
		t.Errorf("%s", errors.Trace(err))
	}
}

func checkDiff(current, desired, add, delete []int) error {
	diff := GetDiff(current, desired)

	if diff.Add == nil {
		diff.Add = []int{}
	}

	if diff.Delete == nil {
		diff.Delete = []int{}
	}

	sort.Ints(diff.Add)
	sort.Ints(diff.Delete)

	if !reflect.DeepEqual(diff.Add, add) {
		return errors.Errorf("diff.add is wrong: expected %v, got %v", add, diff.Add)
	}

	if !reflect.DeepEqual(diff.Delete, delete) {
		return errors.Errorf("diff.delete is wrong: expected %v, got %v", delete, diff.Delete)
	}

	return nil
}
