package utils

import (
	"testing"
	"reflect"
)

func intToInterfaceArr(nums []int) []item{
	arr := make([]item, len(nums))
	for i, v := range nums {
		arr[i] = v
	}
	return arr
}

func TestBasic(t *testing.T) {
	q := NewQ()
	q.Add(1)
	q.Add(2)
	q.Add(3, 4, 5)
	if q.Size() != 5 {
		t.Error("wrong size after adding")
	}

	if !q.Contains(1) {
		t.Error("contain error")
	}

	if num := q.Remove(); num != 1 {
		t.Error("wrong num")
	}

	if q.Contains(1) {
		t.Error("contain error after remove")
	}

	correctNums := []int{2, 3}
	if nums := q.RemoveUpTo(2);  !reflect.DeepEqual(intToInterfaceArr(correctNums), nums) {
		t.Error("remove up to wrong")
	}

	correctNums = []int{4, 5}
	if nums := q.RemoveUpTo(5); !reflect.DeepEqual(intToInterfaceArr(correctNums), nums) {
		t.Log(nums)
		t.Log(correctNums)
		t.Error("remove up to where remove up to input greater than size wrong")
	}

	if q.Size() != 0 {
		t.Error("wrong size after removing")
	}
}

func TestSerialize(t *testing.T) {
	q := NewQ()
	q.Add("1")
	q.Add("2")
	q.Add("3")
	q.Add("4")
	q.Add("5")

	err := q.Serialize("test.txt")
	if err != nil {
		t.Error(err)
	}
	nums1 := q.RemoveUpTo(5)

	q, err = Deserialize("test.txt", 3)
	if err != nil {
		t.Error(err)
	}
	nums2 := q.RemoveUpTo(5)

	if !reflect.DeepEqual(nums1, nums2) {
		t.Log(nums1, nums2)
		t.Error("serialization err")
	}
}
