package dumper

import (
	"io/ioutil"
	"fmt"
	"strings"
	"strconv"
	"os"
	"github.com/regner/albionmarket-client/config"
)

var UnhandledPacketDumper UnhandledPktDumper

const dumpFileName string = "pktDump.go"
const fileHeader string = `package client

`
// Every packet has to have this
const pktStart string = "//+"
// opcode will be stuck at the end of this string, helps with parsing
const pktStartOpcode string = "//"

type packetDumpContainer struct {
	Opcode int16
	DumpString string
}

type packetDumpStorage struct {
	FileHeader string
	DumpContainers []packetDumpContainer
}

type UPDstringParams struct {
	Opcode int16
	parameters []string
}

type UnhandledPktDumper struct {
	DumpedPackets packetDumpStorage
}

func NewUnhandledPktDumper() *UnhandledPktDumper {
	nDumper := UnhandledPktDumper{}
	nDumper.loadDumpedPacketOpcodesFromFile()

	return &nDumper
}

func (params *UPDstringParams) AddParam(paramID uint8, paramType string) {
	var tmp string = "Unknown" + strconv.Itoa(int(paramID)) + " " + paramType + "\t\u0060mapstructure:\"" + strconv.Itoa(int(paramID)) + "\"\u0060"
	params.parameters = append(params.parameters, tmp)
}

func (dumper *UnhandledPktDumper) loadDumpedPacketOpcodesFromFile() {
	if config.GlobalConfiguration.DumpUnknown == false {
		return
	}
	// try to open dumpfile
	data, err := ioutil.ReadFile(dumpFileName)
	if err != nil { // we failed, the file did not exist, so create it
		err = ioutil.WriteFile(dumpFileName, []byte(fileHeader), 0644)
		if err != nil { // we failed again, print the error and bail
			fmt.Println(err)
			return
		}

		// file was created, nothing to be done. this is a virgin file
		return
	}

	// get the data as string from our file
	dataStr := string(data)
	// find first pkt
	startPos := strings.Index(dataStr, pktStart)
	if startPos == -1 { // do nothing if no packet is found
		return
	}

	// get it split up into nice chunks
	p, e := dumper.extractPackets(&dataStr)
	if e != nil { // we failed bail
		return
	}

	dumper.DumpedPackets = *p
}

func (dumper *UnhandledPktDumper) extractPacketDumpContainer(data string) (*packetDumpContainer, error) {
	var pdc packetDumpContainer = packetDumpContainer{}

	data = strings.TrimLeft(data, "\n")
	data = strings.TrimRight(data, "\n")
	var split []string = strings.SplitN(data, "\n", 2)
	v,e := strconv.ParseInt(strings.TrimLeft(split[0], "//"), 10, 16)
	if e != nil {
		return nil, e
	}

	pdc.Opcode = int16(v)
	pdc.DumpString = split[1]

	return &pdc, nil
}

func (dumper *UnhandledPktDumper) extractPackets(data *string) (*packetDumpStorage, error) {
	var pss packetDumpStorage = packetDumpStorage{}

	var pktSplit []string = strings.Split(*data, pktStart)
	pss.FileHeader = pktSplit[0]

	// store our slices
	for i := 1; i < len(pktSplit); i++ {
		//pss.PacketStrings = append(pss.PacketStrings, pktSplit[i])
		pdc, e := dumper.extractPacketDumpContainer(pktSplit[i])
		if e != nil {
			return nil, e
		}
		pss.DumpContainers = append(pss.DumpContainers, *pdc)
	}

	return &pss, nil
}

func (dumper *UnhandledPktDumper) dumpExists(opcode int16) (bool) {
	for _, element := range dumper.DumpedPackets.DumpContainers {
		if element.Opcode == opcode {
			return  true
		}
	}

	return false
}

func (dumper *UnhandledPktDumper) AddPacket(params *UPDstringParams) () {
	if dumper.dumpExists(params.Opcode) {
		return
	}

	var structHeader string = "\n" + pktStart + "\n" + pktStartOpcode + strconv.Itoa(int(params.Opcode)) + "\n"
	var structBody string = "type " + OpcodeEnum[params.Opcode] + " struct {\n"
	for _, e := range params.parameters {
		structBody += e + "\n"
	}
	structBody += "}"

	data := packetDumpContainer{Opcode:params.Opcode, DumpString:structBody}
	dumper.DumpedPackets.DumpContainers = append(dumper.DumpedPackets.DumpContainers, data)

	f, err := os.OpenFile(dumpFileName, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(structHeader + structBody)
	if err != nil {
		panic(err)
	}

	f.Close()
}
