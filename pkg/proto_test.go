package pkg

import (
	"fmt"
	"testing"

	"github.com/wuhuua/fastproto/pkg/parser"
	"github.com/wuhuua/fastproto/pkg/unserializer"
	"google.golang.org/protobuf/proto"
)

func TestParse2Marshall2Unserialize(t *testing.T) {
	parser := parser.NewParser()
	protoSet, err := parser.Load([]string{"../testproto"}, "hello.proto")
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
	fmt.Println("Original Message is:", msg)
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	u := unserializer.NewUnserializer()
	res, err := u.Unserialize(data)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Unserialized Message is:", res)
}
