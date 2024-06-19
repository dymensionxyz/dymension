// Package transfergenesis is the most important piece of the transfer genesis ('genesis bridge') protocol.
//
// The protocol is as follows:
//
//	Rollapps may specify some transfers in THEIR genesis file
//	The transfers will be sent in OnChanOpenConfirm
//	The hub will receive the transfers in any order
//	Before all transfers have arrived, the Hub will reject any OTHER (user submitted) transfers to/from the Rollapp
//	After all transfers have arrived, the Hub will accept any transfers, as usual.
//	Imporant: it is now WRONG to open an ibc connection in the Rollapp->Hub direction.
//	Connections should be opened in the Hub->Rollapp direction only
package transfergenesis
