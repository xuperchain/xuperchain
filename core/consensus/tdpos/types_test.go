package tdpos

import (
	"sort"
	"testing"
)

func TestStable(t *testing.T) {
	termVote0 := &termBallots{
		Address: "address0",
		Ballots: 10,
	}

	termVote1 := &termBallots{
		Address: "address1",
		Ballots: 20,
	}

	termVote2 := &termBallots{
		Address: "address2",
		Ballots: 10,
	}

	termVote3 := &termBallots{
		Address: "address3",
		Ballots: 50,
	}

	termVote4 := &termBallots{
		Address: "address4",
		Ballots: 60,
	}

	testSlice1 := termBallotsSlice{}
	testSlice1 = append(testSlice1, termVote0)
	testSlice1 = append(testSlice1, termVote1)
	testSlice1 = append(testSlice1, termVote2)
	sort.Stable(testSlice1)
	for i := range testSlice1 {
		t.Logf("testSlice1 %v", string(testSlice1[i].Address))
	}

	testSlice2 := termBallotsSlice{}
	testSlice2 = append(testSlice2, termVote1)
	testSlice2 = append(testSlice2, termVote0)
	testSlice2 = append(testSlice2, termVote2)
	sort.Stable(testSlice2)
	for i := range testSlice2 {
		t.Logf("testSlice2 %v", string(testSlice2[i].Address))
	}

	testSlice3 := termBallotsSlice{}
	testSlice3 = append(testSlice3, termVote3)
	testSlice3 = append(testSlice3, termVote0)
	testSlice3 = append(testSlice3, termVote4)
	sort.Stable(testSlice3)
	for i := range testSlice3 {
		t.Logf("testSlice3 %v", string(testSlice3[i].Address))
	}
}
