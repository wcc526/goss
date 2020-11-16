// Copyright 2020 Dean.
// Authors: Dean <dean@csoio.com>
// Date: 2020/9/25 10:25 上午

package goss

import (
	"bytes"
	"encoding/binary"
	"strconv"
	"syscall"
	"unsafe"
)

func Connections(kind string) ([]*Stat, error) {
	var connectionsStat []*Stat
	userEntries, err := BuildUserEntries()
	if err != nil {
		return nil, err
	}
	protocols := netConnectionKindMap[kind]
	for _, protocol := range protocols {
		conn, err := ConnectionsWithProtocol(protocol)
		if err != nil {
			return nil, err
		}
		for _, c := range conn {
			stats := &Stat{
				Proto: netProtocolKindMap[protocol],
				RecvQ: c.RQueue,
				SendQ: c.WQueue,
				Local: &AddrPort{
					Addr: c.SrcIP().String(),
					Port: strconv.Itoa(c.SrcPort()),
				},
				Foreign: &AddrPort{
					Addr: c.DstIP().String(),
					Port: strconv.Itoa(c.DstPort()),
				},
				State: TCPState(c.State).String(),
				Inode: c.Inode,
			}

			if process, ok := userEntries[c.Inode]; ok {
				stats.Process = process
			}
			connectionsStat = append(connectionsStat, stats)
		}
	}
	return connectionsStat, nil
}

func ConnectionsWithProtocol(protocol uint8) ([]*InetDiagMsg, error) {
	hdr := syscall.NlMsghdr{
		Type:  uint16(TCPDIAG_GETSOCK),
		Flags: uint16(syscall.NLM_F_DUMP | syscall.NLM_F_REQUEST),
		Pid:   uint32(0),
	}
	req := InetDiagReq{
		Family: uint8(AF_INET),
		States: AllTCPStates,
	}
	var sizeofInetDiagReq = int(unsafe.Sizeof(InetDiagReq{}))
	byteOrder := GetEndian()
	buf := bytes.NewBuffer(make([]byte, sizeofInetDiagReq))
	buf.Reset()
	if err := binary.Write(buf, byteOrder, req); err != nil {
		// This never returns an error.
		return nil, err
	}
	b := buf.Bytes()
	req2 := syscall.NetlinkMessage{Header: hdr, Data: b}
	return NetlinkInetDiag(req2)
}
