package mount

import "testing"

func TestMountInfo(t *testing.T) {
	infos, err := Infos()
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) == 0 {
		t.Errorf("expected some amount of mountinfo, but got none")
	}
}
