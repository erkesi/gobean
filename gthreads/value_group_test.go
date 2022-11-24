package gthreads

import (
	"github.com/erkesi/gobean/grecovers"
	"strings"
	"testing"
)

func TestGo(t *testing.T) {
	var vg ValueGroup
	vg.SetLimit(3)
	for i := 0; i < 100; i++ {
		tmpI := i
		vg.Go(grecovers.RecoverVGFn(func() (interface{}, error) {
			return tmpI, nil
		}))
	}

	res, err := vg.Wait()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("vals:%v", res)
	if len(res) != 100 || res[0].(int) != 0 || res[len(res)-1].(int) != 99 {
		t.Fatalf("len:%d", len(res))
	}
}

func TestGoForErr(t *testing.T) {
	var vg ValueGroup

	for i := 0; i < 100; i++ {
		tmpI := i
		vg.Go(grecovers.RecoverVGFn(func() (interface{}, error) {
			if tmpI == 10 {
				panic("err 10")
			}
			return tmpI, nil
		}))
	}

	res, err := vg.Wait()
	t.Logf("len vals:%v", len(res))
	t.Logf("len err:%v", err)
	if !strings.Contains(err.Error(), "panic") {
		t.Fatal(err)
	}
}
