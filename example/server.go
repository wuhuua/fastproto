package example

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type HelloServer interface {
	Hello(context.Context, *string) (*string, error)
}

type Hello struct {
	HelloServer
}

func (s *Hello) Hello(ctx context.Context, req *string) (*string, error) {
	log.Println("Received request:", *req)
	res := "testdata"
	return &res, nil
}

var _Hello_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Hello",
	HandlerType: (*HelloServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hello",
			Handler:    _Hello_Hello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hello.proto",
}

func _Hello_Hello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(string)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*Hello).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Hello/Hello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*Hello).Hello(ctx, req.(*string))
	}
	return interceptor(ctx, in, info, handler)
}

func Run() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	RegisterHelloServer(s, &Hello{})

	log.Println("gRPC server listening on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func RegisterHelloServer(s *grpc.Server, h *Hello) {
	s.RegisterService(&_Hello_serviceDesc, h)
}
