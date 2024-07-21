package parser

import (
	"fmt"
	"testing"
)

func TestParse2ProtoSet(t *testing.T) {
	parser := NewParser()
	protoSet, err := parser.Load([]string{"testproto"}, "hello.proto")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(protoSet)
}

func TestNewMessage(t *testing.T) {
	parser := NewParser()
	protoSet, err := parser.Load([]string{"../../testproto"}, "hello.proto")
	if err != nil {
		t.Fatal(err)
	}
	testMsg := `{
		"name": "fastproto",
		"msg": "hello"
	}`
	msg, err := protoSet.NewMessage("/testproto.HelloService/Hello", testMsg)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(msg)
}
