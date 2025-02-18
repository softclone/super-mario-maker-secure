package types

import "github.com/PretendoNetwork/nex-go"

type CourseMetadata struct {
	DataID      uint64
	OwnerPID    uint32
	Size        uint32
	CreatedTime *nex.DateTime
	UpdatedTime *nex.DateTime
	Name        string
	MetaBinary  []byte
	Stars       uint32
	Attempts    uint32
	Failures    uint32
	Completions uint32
	Flag        uint32
	DataType    uint16
	Period      uint16
}
