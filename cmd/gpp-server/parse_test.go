package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"net/url"
	"os"
	"testing"
)

func TestParser(t *testing.T) {
	fp, err := os.Open("config.json")
	if err != nil {
		t.Fatalf("open failed: %s", err)
	}

	defer fp.Close()

	var _cfg CFG
	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		t.Errorf("read error: %s", err)
	}

	err = json.Unmarshal(buf, &_cfg)
	if err != nil {
		t.Errorf("parser json error: %s", err)
	}

	fmt.Printf("%+v\n", _cfg)
}
