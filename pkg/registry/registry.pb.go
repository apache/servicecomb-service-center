package registry

import (
	"github.com/gogo/protobuf/proto"
)

func (m *Schema) Reset()         { *m = Schema{} }
func (m *Schema) String() string { return proto.CompactTextString(m) }
func (*Schema) ProtoMessage()    {}

func (m *MicroService) Reset()         { *m = MicroService{} }
func (m *MicroService) String() string { return proto.CompactTextString(m) }
func (*MicroService) ProtoMessage()    {}

func (m *MicroServiceInstance) Reset()         { *m = MicroServiceInstance{} }
func (m *MicroServiceInstance) String() string { return proto.CompactTextString(m) }
func (*MicroServiceInstance) ProtoMessage()    {}
