// Package transfergenesis is the most important piece of the transfer genesis ('genesis bridge') protocol.
//
// The protocol is as follows:
//
//			Rollapps may specify some transfers in THEIR genesis file
//			The transfers will be sent in OnChanOpenConfirm with a 'special' memo
//			The hub will receive the special transfers in any order
//	        Before the arrival of a normal transfer, the Hub will block transfers Hub->RA
//	        After the arrival of the first normal transfer, the Hub will allow transfers Hub->RA and will not allow more genesis transfers
//
//			Important: it is now WRONG to open an ibc connection in the Rollapp->Hub direction.
//			Connections should be opened in the Hub->Rollapp direction only
package transfergenesis
