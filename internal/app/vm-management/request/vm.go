package request

type Request struct {
	Command      string
	VmName       string
	Cpu          uint
	Ram          uint
	SourceVmName string
	DestVmName   string
	Input        string
	OriginVm     string
	OriginPath   string
	DestVm       string
	DestPath     string
}
