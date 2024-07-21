package parser

import (
	"errors"
	"fmt"
	"os"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Parser struct {
}

type MethodInfo struct {
	Package    string
	Service    string
	FullMethod string
}

type ProtoSet struct {
	methods []MethodInfo
	mds     map[string]protoreflect.MethodDescriptor
}

func NewParser() *Parser {
	return &Parser{}
}

func (c *Parser) Load(importPaths []string, filenames ...string) (*ProtoSet, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// 没有提供proto引用目录时使用当前目录
	if len(importPaths) == 0 {
		importPaths = append(importPaths, dir)
	}

	parser := protoparse.Parser{
		ImportPaths:      importPaths,
		InferImportPaths: false,
	}

	fds, err := parser.ParseFiles(filenames...)
	if err != nil {
		return nil, err
	}
	fdset := &descriptorpb.FileDescriptorSet{}

	seen := make(map[string]struct{})
	for _, fd := range fds {
		fdset.File = append(fdset.File, walkFileDescriptors(seen, fd)...)
	}
	return c.convertToMethodInfo(fdset)
}

func (c *Parser) convertToMethodInfo(fdset *descriptorpb.FileDescriptorSet) (*ProtoSet, error) {
	files, err := protodesc.NewFiles(fdset)
	if err != nil {
		return nil, err
	}
	protoSet := &ProtoSet{}
	var rtn []MethodInfo
	protoSet.mds = make(map[string]protoreflect.MethodDescriptor)
	appendMethodInfo := func(
		fd protoreflect.FileDescriptor,
		sd protoreflect.ServiceDescriptor,
		md protoreflect.MethodDescriptor,
	) {
		name := fmt.Sprintf("/%s/%s", sd.FullName(), md.Name())
		protoSet.mds[name] = md
		rtn = append(rtn, MethodInfo{
			Package:    string(fd.Package()),
			Service:    string(sd.Name()),
			FullMethod: name,
		})
	}
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		sds := fd.Services()
		for i := 0; i < sds.Len(); i++ {
			sd := sds.Get(i)
			mds := sd.Methods()
			for j := 0; j < mds.Len(); j++ {
				md := mds.Get(j)
				appendMethodInfo(fd, sd, md)
			}
		}

		messages := fd.Messages()

		stack := make([]protoreflect.MessageDescriptor, 0, messages.Len())
		for i := 0; i < messages.Len(); i++ {
			stack = append(stack, messages.Get(i))
		}

		for len(stack) > 0 {
			message := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			_, errFind := protoregistry.GlobalTypes.FindMessageByName(message.FullName())
			if errors.Is(errFind, protoregistry.NotFound) {
				err = protoregistry.GlobalTypes.RegisterMessage(dynamicpb.NewMessageType(message))
				if err != nil {
					return false
				}
			}

			nested := message.Messages()
			for i := 0; i < nested.Len(); i++ {
				stack = append(stack, nested.Get(i))
			}
		}

		return true
	})
	if err != nil {
		return nil, err
	}
	protoSet.methods = rtn
	return protoSet, nil
}

func walkFileDescriptors(seen map[string]struct{}, fd *desc.FileDescriptor) []*descriptorpb.FileDescriptorProto {
	fds := []*descriptorpb.FileDescriptorProto{}

	if _, ok := seen[fd.GetName()]; ok {
		return fds
	}
	seen[fd.GetName()] = struct{}{}
	fds = append(fds, fd.AsFileDescriptorProto())

	for _, dep := range fd.GetDependencies() {
		deps := walkFileDescriptors(seen, dep)
		fds = append(fds, deps...)
	}

	return fds
}

func (c *ProtoSet) GetMethodDescriptor(method string) protoreflect.MethodDescriptor {
	return c.mds[method]
}

func (c *ProtoSet) NewMessage(fullMethod string, jsonStr string) (*dynamicpb.Message, error) {
	md := c.GetMethodDescriptor(fullMethod)
	if md == nil {
		return nil, fmt.Errorf("method not found: %s", fullMethod)
	}
	msgType := md.Input()
	msg := dynamicpb.NewMessage(msgType)
	err := protojson.Unmarshal([]byte(jsonStr), msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
