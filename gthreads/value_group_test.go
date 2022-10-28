package gthreads

import (
	"github.com/erkesi/gobean/grecovers"
	"strings"
	"testing"
)

func TestGo(t *testing.T) {
	var vg ValueGroup

	for i:=0;i<100;i++{
		tmpI:=i
		vg.Go(grecovers.RecoverFnWithVG(func() (interface{}, error) {
			return tmpI,nil
		}))
	}

	res, err := vg.Wait()
	if err!=nil{
		t.Fatal(err)
	}

	t.Logf("res:%v", res)
	if len(res)!=100 {
		t.Fatalf("len:%d", len(res))
	}
}

func TestGoForErr(t *testing.T) {
	var vg ValueGroup
	
	for i:=0;i<100;i++{
		tmpI:=i
		vg.Go(grecovers.RecoverFnWithVG(func() (interface{}, error) {
			if tmpI==10 {
				panic("err 10")
			}
			return tmpI,nil
		}))
	}

	res, err := vg.Wait()
	t.Logf("len res:%v", len(res))
	if !strings.Contains(err.Error(),"panic"){
		t.Fatal(err)
	}
}