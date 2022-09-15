package protoext

import (
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// https://github.com/protocolbuffers/protobuf-go/blob/master/encoding/protojson/well_known_types.go
var wellKnownMessages = []proto.Message{
	&anypb.Any{},

	&timestamppb.Timestamp{},

	&durationpb.Duration{},

	&wrapperspb.BoolValue{},
	&wrapperspb.Int32Value{},
	&wrapperspb.Int64Value{},
	&wrapperspb.UInt32Value{},
	&wrapperspb.UInt64Value{},
	&wrapperspb.FloatValue{},
	&wrapperspb.DoubleValue{},
	&wrapperspb.StringValue{},
	&wrapperspb.BytesValue{},

	&structpb.Struct{},

	&structpb.ListValue{},

	&structpb.Value{},

	&fieldmaskpb.FieldMask{},

	&emptypb.Empty{},
}
var wellKnownPtrTypes = map[reflect2.Type]bool{}

func init() {
	for _, m := range wellKnownMessages {
		typ := reflect2.TypeOf(m)
		wellKnownPtrTypes[typ] = true
	}
}
