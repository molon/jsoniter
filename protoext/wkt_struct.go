package protoext

// var wktStructCodec = NewElemTypeCodec(
// 	func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
// 		x := ((*structpb.Struct)(ptr))
// 		stream.WriteVal(x.Fields)
// 	},
// 	func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
// 		iter.ReadMapCB(func(*jsoniter.Iterator, string) bool{

// 		})
// 		err := unmarshalValue(((*structpb.Value)(ptr)), iter)

// 		if err != nil {
// 			iter.ReportError("protobuf", fmt.Sprintf("%s: %v", Value_message_fullname, err))
// 		}
// 	},
// )
