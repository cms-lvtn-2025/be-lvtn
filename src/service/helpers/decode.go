package helpers

import (
	"encoding/base64"

	"google.golang.org/protobuf/types/known/structpb"
)

// Hàm đệ quy decode Base64 trong structpb.Struct
func DecodeStructpb(s *structpb.Struct) {
	for k, v := range s.Fields {
		decodeValue(v)
		s.Fields[k] = v
	}
}

func decodeValue(v *structpb.Value) {
	switch kind := v.Kind.(type) {
	case *structpb.Value_StringValue:
		decoded, err := base64.StdEncoding.DecodeString(kind.StringValue)
		if err == nil {
			v.Kind = &structpb.Value_StringValue{StringValue: string(decoded)}
		}
	case *structpb.Value_StructValue:
		DecodeStructpb(kind.StructValue)
	case *structpb.Value_ListValue:
		for i, item := range kind.ListValue.Values {
			decodeValue(item)
			kind.ListValue.Values[i] = item
		}
	}
}
