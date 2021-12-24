package network

//General Packet Processor interface.
//Implement to receive packets from parsers
type PacketProcessor interface {
	ProcessPacket(pkt *Packet) error
}
